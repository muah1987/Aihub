package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/muah1987/Aihub/internal/activity"
	"github.com/muah1987/Aihub/internal/agent"
	"github.com/muah1987/Aihub/internal/integration"
	"github.com/muah1987/Aihub/internal/knowledge"
	"github.com/muah1987/Aihub/internal/analytics"
	"github.com/muah1987/Aihub/internal/auth"
	"github.com/muah1987/Aihub/internal/chat"
	"github.com/muah1987/Aihub/internal/config"
	"github.com/muah1987/Aihub/internal/database"
	"github.com/muah1987/Aihub/internal/deployment"
	"github.com/muah1987/Aihub/internal/email"
	"github.com/muah1987/Aihub/internal/memory"
	"github.com/muah1987/Aihub/internal/monitoring"
	"github.com/muah1987/Aihub/internal/notification"
	"github.com/muah1987/Aihub/internal/organization"
	"github.com/muah1987/Aihub/internal/project"
	"github.com/muah1987/Aihub/internal/provider"
	"github.com/muah1987/Aihub/internal/rbac"
	"github.com/muah1987/Aihub/internal/router"
	"github.com/muah1987/Aihub/internal/scheduler"
	"github.com/muah1987/Aihub/internal/team"
	"github.com/muah1987/Aihub/internal/terminal"
	"github.com/muah1987/Aihub/internal/tools"
	"github.com/muah1987/Aihub/internal/webhook"
	"github.com/muah1987/Aihub/internal/workflow"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting Aihub server...")

	cfg := config.Load()
	log.Printf("Environment: %s", cfg.Server.Environment)

	db, err := database.Connect(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("Connected to database")

	if err := database.RunMigrations(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	emailService := email.NewService(&email.Config{
		Host:     cfg.SMTP.Host,
		Port:     cfg.SMTP.Port,
		Username: cfg.SMTP.Username,
		Password: cfg.SMTP.Password,
		From:     cfg.SMTP.From,
		BaseURL:  cfg.SMTP.BaseURL,
	})

	jwtService := auth.NewJWTService(&cfg.JWT)
	authService := auth.NewService(db, jwtService, emailService)

	providerService, err := provider.NewService(db, cfg.Encryption.Key)
	if err != nil {
		log.Fatalf("Failed to initialize provider service: %v", err)
	}

	projectService := project.NewService(db, providerService)

	chatHub := chat.NewHub()
	go chatHub.Run()
	chatService := chat.NewService(db, chatHub)

	cpuLimit, _ := strconv.ParseFloat(cfg.Docker.CPULimit, 64)
	containerMgr, err := terminal.NewContainerManager(
		cfg.Docker.Host, cfg.Docker.SandboxImage, cfg.Docker.MemoryLimit, cpuLimit,
	)
	if err != nil {
		log.Printf("Warning: Docker unavailable: %v — terminal features disabled", err)
		containerMgr = nil
	}

	var terminalService *terminal.Service
	if containerMgr != nil {
		terminalService = terminal.NewService(db, containerMgr)
	}

	// Phase 2
	memoryService := memory.NewService(db)
	rbacService := rbac.NewService(db)
	orgService := organization.NewService(db, emailService)
	agentService := agent.NewService(db, providerService, chatService, memoryService)
	teamService := team.NewService(db)
	orchestrator := team.NewOrchestrator(db, agentService, chatService, memoryService, providerService)

	// Phase 3
	deploymentService, err := deployment.NewService(db, cfg.Encryption.Key)
	if err != nil {
		log.Fatalf("Failed to initialize deployment service: %v", err)
	}

	// Phase 4
	notifyHub := notification.NewHub()
	notifyService := notification.NewService(db, notifyHub)
	webhookService := webhook.NewService(db)
	monitoringService := monitoring.NewService(db, deploymentService)

	// Phase 5
	toolsService := tools.NewService(db)
	analyticsService := analytics.NewService(db)
	agentService.SetUsageRecorder(analyticsService)
	agentService.SetToolProvider(toolsService)

	// Phase 6
	workflowService := workflow.NewService(db)
	schedulerService := scheduler.NewService(db)
	activityService := activity.NewService(db)

	// Phase 7
	knowledgeService := knowledge.NewService(db)

	// Phase 8
	integrationService := integration.NewService(db)

	// Handlers
	handlers := &router.Handlers{
		Auth:         auth.NewHandler(authService),
		Provider:     provider.NewHandler(providerService),
		Project:      project.NewHandler(projectService),
		Chat:         chat.NewHandler(chatService, chatHub),
		Agent:        agent.NewHandler(agentService),
		Organization: organization.NewHandler(orgService, rbacService),
		Memory:       memory.NewHandler(memoryService),
		Team:         team.NewHandler(teamService, orchestrator),
		Deployment:   deployment.NewHandler(deploymentService),
		Webhook:      webhook.NewHandler(webhookService, deploymentService, notifyService),
		Monitoring:   monitoring.NewHandler(monitoringService),
		Notification: notification.NewHandler(notifyService, notifyHub),
		Tools:        tools.NewHandler(toolsService),
		Analytics:    analytics.NewHandler(analyticsService),
		Workflow:     workflow.NewHandler(workflowService),
		Scheduler:    scheduler.NewHandler(schedulerService),
		Activity:     activity.NewHandler(activityService),
		Knowledge:    knowledge.NewHandler(knowledgeService),
		Integration:  integration.NewHandler(integrationService),
	}

	if terminalService != nil {
		handlers.Terminal = terminal.NewHandler(terminalService)
	}

	handler := router.New(handlers, jwtService, cfg.CORS.AllowedOrigins)

	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("Shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if containerMgr != nil {
			containerMgr.Close()
		}
		if err := srv.Shutdown(ctx); err != nil {
			log.Fatalf("Server shutdown failed: %v", err)
		}
	}()

	log.Printf("Server listening on %s", addr)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("Server failed: %v", err)
	}
	log.Println("Server stopped")
}
