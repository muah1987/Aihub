package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/muah1987/Aihub/internal/agent"
	"github.com/muah1987/Aihub/internal/analytics"
	"github.com/muah1987/Aihub/internal/auth"
	"github.com/muah1987/Aihub/internal/chat"
	"github.com/muah1987/Aihub/internal/deployment"
	"github.com/muah1987/Aihub/internal/memory"
	"github.com/muah1987/Aihub/internal/middleware"
	"github.com/muah1987/Aihub/internal/monitoring"
	"github.com/muah1987/Aihub/internal/notification"
	"github.com/muah1987/Aihub/internal/organization"
	"github.com/muah1987/Aihub/internal/project"
	"github.com/muah1987/Aihub/internal/provider"
	"github.com/muah1987/Aihub/internal/team"
	"github.com/muah1987/Aihub/internal/terminal"
	"github.com/muah1987/Aihub/internal/tools"
	"github.com/muah1987/Aihub/internal/webhook"
)

type Handlers struct {
	Auth         *auth.Handler
	Provider     *provider.Handler
	Project      *project.Handler
	Chat         *chat.Handler
	Terminal     *terminal.Handler
	Agent        *agent.Handler
	Organization *organization.Handler
	Memory       *memory.Handler
	Team         *team.Handler
	Deployment   *deployment.Handler
	Webhook      *webhook.Handler
	Monitoring   *monitoring.Handler
	Notification *notification.Handler
	Tools        *tools.Handler
	Analytics    *analytics.Handler
}

func New(
	handlers *Handlers,
	jwtService *auth.JWTService,
	corsOrigins string,
) http.Handler {
	r := chi.NewRouter()

	// Global middleware
	rateLimiter := middleware.NewRateLimiter(10, 50)
	r.Use(middleware.CORS(corsOrigins))
	r.Use(middleware.Logger)
	r.Use(rateLimiter.Middleware)
	r.Use(auth.InjectJWTService(jwtService))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Public GitHub webhook ingress (no auth, HMAC-verified)
	if handlers.Webhook != nil {
		r.Post("/webhooks/github/{webhookId}", handlers.Webhook.GitHubIncoming)
	}

	// API v1
	r.Route("/api/v1", func(r chi.Router) {
		// Auth routes (public)
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", handlers.Auth.Register)
			r.Post("/login", handlers.Auth.Login)
			r.Post("/refresh", handlers.Auth.Refresh)
			r.Post("/verify-email", handlers.Auth.VerifyEmail)
			r.Post("/2fa/verify-login", handlers.Auth.LoginVerify2FA)

			// Protected auth routes
			r.Group(func(r chi.Router) {
				r.Use(auth.Middleware(jwtService))
				r.Get("/me", handlers.Auth.Me)
				r.Post("/logout", handlers.Auth.Logout)
				r.Post("/resend-verification", handlers.Auth.ResendVerification)
				r.Post("/2fa/setup", handlers.Auth.Setup2FA)
				r.Post("/2fa/confirm", handlers.Auth.Confirm2FA)
				r.Post("/2fa/disable", handlers.Auth.Disable2FA)
			})
		})

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(auth.Middleware(jwtService))

			// Notifications
			if handlers.Notification != nil {
				r.Route("/notifications", func(r chi.Router) {
					r.Get("/", handlers.Notification.List)
					r.Post("/read-all", handlers.Notification.MarkAllRead)
					r.Post("/{id}/read", handlers.Notification.MarkRead)
					r.Delete("/{id}", handlers.Notification.Delete)
				})
			}

			// Provider connections
			r.Route("/providers", func(r chi.Router) {
				r.Get("/", handlers.Provider.List)
				r.Post("/", handlers.Provider.Create)
				r.Delete("/{id}", handlers.Provider.Delete)
				r.Post("/{id}/validate", handlers.Provider.Validate)
			})

			// Organizations
			if handlers.Organization != nil {
				r.Route("/organizations", func(r chi.Router) {
					r.Get("/", handlers.Organization.List)
					r.Post("/", handlers.Organization.Create)
					r.Route("/{orgId}", func(r chi.Router) {
						r.Get("/", handlers.Organization.Get)
						r.Put("/", handlers.Organization.Update)
						r.Delete("/", handlers.Organization.Delete)
						r.Get("/members", handlers.Organization.ListMembers)
						r.Post("/invite", handlers.Organization.InviteMember)
						r.Delete("/members/{userId}", handlers.Organization.RemoveMember)
						r.Put("/members/{userId}/role", handlers.Organization.UpdateMemberRole)
					})
				})
				r.Post("/invitations/{token}/accept", handlers.Organization.AcceptInvitation)
			}

			// Projects
			r.Route("/projects", func(r chi.Router) {
				r.Get("/", handlers.Project.List)
				r.Post("/", handlers.Project.Create)
				r.Get("/repos", handlers.Project.ListRepos)

				r.Route("/{id}", func(r chi.Router) {
					r.Get("/", handlers.Project.Get)
					r.Put("/", handlers.Project.Update)
					r.Delete("/", handlers.Project.Delete)

					// Chat
					r.Get("/messages", handlers.Chat.GetHistory)
					r.Post("/messages", handlers.Chat.PostMessage)

					// Agents
					r.Route("/agents", func(r chi.Router) {
						r.Get("/", handlers.Agent.List)
						r.Post("/", handlers.Agent.Create)
						r.Put("/{agentId}", handlers.Agent.Update)
						r.Delete("/{agentId}", handlers.Agent.Delete)
						r.Post("/{agentId}/invoke", handlers.Agent.Invoke)

						// Agent tool bindings
						if handlers.Tools != nil {
							r.Get("/{agentId}/tools", handlers.Tools.ListBindings)
							r.Post("/{agentId}/tools", handlers.Tools.BindTool)
							r.Delete("/{agentId}/tools/{toolId}", handlers.Tools.UnbindTool)
							r.Post("/{agentId}/tools/{toolId}/execute", handlers.Tools.ExecuteTool)
						}
					})

					// Project-level tools
					if handlers.Tools != nil {
						r.Route("/tools", func(r chi.Router) {
							r.Get("/", handlers.Tools.ListTools)
							r.Post("/", handlers.Tools.CreateTool)
							r.Put("/{toolId}", handlers.Tools.UpdateTool)
							r.Delete("/{toolId}", handlers.Tools.DeleteTool)
							r.Get("/executions", handlers.Tools.ListExecutions)
						})
					}

					// Analytics
					if handlers.Analytics != nil {
						r.Route("/analytics", func(r chi.Router) {
							r.Get("/summary", handlers.Analytics.Summary)
							r.Get("/daily", handlers.Analytics.Daily)
							r.Get("/models", handlers.Analytics.Models)
							r.Get("/recent", handlers.Analytics.Recent)
							r.Get("/budget", handlers.Analytics.GetBudget)
							r.Post("/budget", handlers.Analytics.SetBudget)
						})
					}

					// Terminal sessions
					r.Route("/terminal", func(r chi.Router) {
						r.Get("/sessions", handlers.Terminal.ListSessions)
						r.Post("/sessions", handlers.Terminal.CreateSession)
						r.Delete("/sessions/{sessionId}", handlers.Terminal.StopSession)
					})

					// Team Memory
					if handlers.Memory != nil {
						r.Route("/memory", func(r chi.Router) {
							r.Get("/", handlers.Memory.List)
							r.Post("/", handlers.Memory.Set)
							r.Get("/search", handlers.Memory.Search)
							r.Get("/{category}/{key}", handlers.Memory.Get)
							r.Delete("/{memoryId}", handlers.Memory.Delete)
						})
					}

					// Agent Teams
					if handlers.Team != nil {
						r.Route("/teams", func(r chi.Router) {
							r.Get("/", handlers.Team.ListTeams)
							r.Post("/", handlers.Team.CreateTeam)
							r.Route("/{teamId}", func(r chi.Router) {
								r.Get("/", handlers.Team.GetTeam)
								r.Put("/", handlers.Team.UpdateTeam)
								r.Delete("/", handlers.Team.DeleteTeam)
								r.Post("/invoke", handlers.Team.InvokeTeam)
								r.Get("/members", handlers.Team.ListMembers)
								r.Post("/members", handlers.Team.AddMember)
								r.Delete("/members/{agentId}", handlers.Team.RemoveMember)
								r.Put("/leader", handlers.Team.SetLeader)
								r.Get("/tasks", handlers.Team.ListTasks)
								r.Get("/tasks/{taskId}", handlers.Team.GetTask)
							})
						})
					}

					// Deployment (env vars + VPS targets + runs)
					if handlers.Deployment != nil {
						r.Route("/env", func(r chi.Router) {
							r.Get("/", handlers.Deployment.ListEnvVars)
							r.Post("/", handlers.Deployment.SetEnvVar)
							r.Delete("/{envId}", handlers.Deployment.DeleteEnvVar)
						})

						r.Route("/deploy", func(r chi.Router) {
							r.Get("/", handlers.Deployment.ListTargets)
							r.Post("/", handlers.Deployment.CreateTarget)
							r.Get("/runs", handlers.Deployment.ListRuns)
							r.Route("/{targetId}", func(r chi.Router) {
								r.Put("/", handlers.Deployment.UpdateTarget)
								r.Delete("/", handlers.Deployment.DeleteTarget)
								r.Post("/trigger", handlers.Deployment.Deploy)
								r.Get("/runs", handlers.Deployment.ListRuns)

								// Monitoring per target
								if handlers.Monitoring != nil {
									r.Post("/metrics/collect", handlers.Monitoring.CollectMetrics)
									r.Get("/metrics/latest", handlers.Monitoring.LatestMetric)
									r.Get("/metrics/history", handlers.Monitoring.MetricsHistory)
									r.Get("/stages", handlers.Monitoring.ListStages)
									r.Post("/stages", handlers.Monitoring.CreateStage)
									r.Delete("/stages/{stageId}", handlers.Monitoring.DeleteStage)
								}
							})
						})

						r.Get("/runs/{runId}", handlers.Deployment.GetRun)
					}

					// Webhooks
					if handlers.Webhook != nil {
						r.Route("/webhooks", func(r chi.Router) {
							r.Get("/", handlers.Webhook.List)
							r.Post("/", handlers.Webhook.Create)
							r.Delete("/{webhookId}", handlers.Webhook.Delete)
							r.Put("/{webhookId}/active", handlers.Webhook.SetActive)
							r.Get("/{webhookId}/secret", handlers.Webhook.GetSecret)
						})
					}
				})
			})
		})

		// WebSocket routes (auth via query param)
		r.Get("/projects/{id}/chat/ws", handlers.Chat.WebSocket)
		r.Get("/projects/{id}/terminal/ws/{sessionId}", handlers.Terminal.WebSocket)

		// Notification WebSocket
		if handlers.Notification != nil {
			r.Get("/notifications/ws", handlers.Notification.WebSocket)
		}
	})

	return r
}
