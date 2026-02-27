package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/muah1987/Aihub/internal/agent"
	"github.com/muah1987/Aihub/internal/auth"
	"github.com/muah1987/Aihub/internal/chat"
	"github.com/muah1987/Aihub/internal/memory"
	"github.com/muah1987/Aihub/internal/middleware"
	"github.com/muah1987/Aihub/internal/organization"
	"github.com/muah1987/Aihub/internal/project"
	"github.com/muah1987/Aihub/internal/provider"
	"github.com/muah1987/Aihub/internal/team"
	"github.com/muah1987/Aihub/internal/terminal"
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
					})

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
				})
			})
		})

		// WebSocket routes (auth via query param)
		r.Get("/projects/{id}/chat/ws", handlers.Chat.WebSocket)
		r.Get("/projects/{id}/terminal/ws/{sessionId}", handlers.Terminal.WebSocket)
	})

	return r
}
