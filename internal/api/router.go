package api

import (
	"github.com/gin-gonic/gin"

	"ymmo/internal/handler"
	"ymmo/internal/middleware"
	"ymmo/internal/service"
)

func NewRouter(
	authH *handler.AuthHandler,
	propertyH *handler.PropertyHandler,
	statsH *handler.StatsHandler,
	adminH *handler.AdminHandler,
	agentH *handler.AgentHandler,
	messageH *handler.MessageHandler,
	authSvc service.AuthService,
) *gin.Engine {

	r := gin.Default()

	r.Use(middleware.CORS())
	r.Use(middleware.Logger())
	r.Static("/uploads", "./uploads")

	v1 := r.Group("/api/v1")
	{
		// ── Auth ──────────────────────────────────────────────
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authH.Register)
			auth.POST("/login", authH.Login)

			// NOTE: Group("") (et non "/") pour éviter un slash final
			// qui casserait les routes enregistrées avec un chemin vide.
			authRequired := auth.Group("")
			authRequired.Use(middleware.JWTAuth(authSvc))
			{
				authRequired.POST("/logout", authH.Logout)
				authRequired.GET("/me", authH.Me)
			}
		}

		// ── Biens ─────────────────────────────────────────────
		properties := v1.Group("/properties")
		{
			properties.GET("", propertyH.List)
			properties.GET("/:id", propertyH.GetByID)

			// IMPORTANT: Group("") et non Group("/")
			// Group("/") créerait la route "/api/v1/properties/" (avec slash final)
			// pour .POST(""), alors que le frontend appelle "/api/v1/properties"
			// (sans slash). Gin redirigerait (308) vers la version avec slash,
			// mais cette redirection interne ne passe pas par le middleware CORS
			// → le navigateur bloque la redirection avec une fausse erreur CORS.
			agentOnly := properties.Group("")
			agentOnly.Use(middleware.JWTAuth(authSvc), middleware.RequireRole("agent", "admin"))
			{
				agentOnly.POST("", propertyH.Create)
				agentOnly.PUT("/:id", propertyH.Update)
				agentOnly.DELETE("/:id", propertyH.Delete)
				agentOnly.POST("/:id/images", propertyH.UploadImages)
			}

			contactRoute := properties.Group("")
			contactRoute.Use(middleware.JWTAuth(authSvc))
			{
				contactRoute.POST("/:id/contact", agentH.ContactAgent)
			}
		}

		// ── Agent ─────────────────────────────────────────────
		agent := v1.Group("/agent")
		agent.Use(middleware.JWTAuth(authSvc), middleware.RequireRole("agent", "admin"))
		{
			agent.GET("/messages", agentH.GetMessages)
		}

		// ── Admin ─────────────────────────────────────────────
		admin := v1.Group("/admin")
		admin.Use(middleware.JWTAuth(authSvc), middleware.RequireRole("admin"))
		{
			admin.GET("/users", adminH.ListUsers)
			admin.PATCH("/users/:id/role", adminH.UpdateRole)
			admin.DELETE("/users/:id", adminH.DeleteUser)
		}

		// ── Messages (conversations bidirectionnelles) ─────────
		// Accessible à TOUT utilisateur authentifié (client OU agent) :
		// le service détermine le rôle via la conversation elle-même.
		messages := v1.Group("/messages")
		messages.Use(middleware.JWTAuth(authSvc))
		{
			messages.GET("", messageH.GetMyMessages)
			messages.POST("/:id/reply", messageH.Reply)
			messages.PATCH("/:id/read", messageH.MarkAsRead)
		}

		// ── Stats ─────────────────────────────────────────────
		stats := v1.Group("/stats")
		{
			stats.GET("/market", statsH.Market)
			stats.GET("/popular", statsH.Popular)

			statsAuth := stats.Group("")
			statsAuth.Use(middleware.JWTAuth(authSvc), middleware.RequireRole("agent", "admin"))
			{
				statsAuth.GET("/dashboard", statsH.Dashboard)
			}
		}
	}

	return r
}
