package api

import (
	"github.com/gin-gonic/gin"

	"ymmo/internal/handler"
	"ymmo/internal/middleware"
)

// NewRouter initialise toutes les routes de l'application.
func NewRouter(
	authH *handler.AuthHandler,
	propertyH *handler.PropertyHandler,
	statsH *handler.StatsHandler,
	adminH *handler.AdminHandler,
	agentH *handler.AgentHandler,
) *gin.Engine {

	r := gin.Default()

	// Middlewares globaux
	//Un middleware est un composant logiciel qui s’intercale entre deux parties d’une application pour traiter les requêtes ou les réponses.
	r.Use(middleware.CORS())
	r.Use(middleware.Logger())

	// Fichiers statiques (images uploadées)
	r.Static("/uploads", "./uploads")

	v1 := r.Group("/api/v1")
	{
		//  Auth (public)
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authH.Register)
			auth.POST("/login", authH.Login)

			// Ces deux routes nécessitent un token valide
			authRequired := auth.Group("/")
			authRequired.Use(middleware.JWTAuth())
			{
				authRequired.POST("/logout", authH.Logout)
				authRequired.GET("/me", authH.Me)
			}
		}

		// Biens immobiliers
		properties := v1.Group("/properties")
		{
			// Routes publiques
			properties.GET("", propertyH.List) // ?city=&type=&min_price=&max_price=&page=
			properties.GET("/:id", propertyH.GetByID)

			// Routes agent uniquement
			agentOnly := properties.Group("/")
			agentOnly.Use(middleware.JWTAuth(), middleware.RequireRole("agent", "admin"))
			{
				agentOnly.POST("", propertyH.Create)
				agentOnly.PUT("/:id", propertyH.Update)
				agentOnly.DELETE("/:id", propertyH.Delete)
				agentOnly.POST("/:id/images", propertyH.UploadImages)
			}

			// Contact agent (client connecté)
			contactRoute := properties.Group("/")
			contactRoute.Use(middleware.JWTAuth())
			{
				contactRoute.POST("/:id/contact", agentH.ContactAgent)
			}
		}

		//  Agent
		agent := v1.Group("/agent")
		agent.Use(middleware.JWTAuth(), middleware.RequireRole("agent", "admin"))
		{
			agent.GET("/messages", agentH.GetMessages)
		}

		//  Admin
		admin := v1.Group("/admin")
		admin.Use(middleware.JWTAuth(), middleware.RequireRole("admin"))
		{
			admin.GET("/users", adminH.ListUsers)
			admin.PATCH("/users/:id/role", adminH.UpdateRole)
			admin.DELETE("/users/:id", adminH.DeleteUser)
		}

		// Stats & data (publiques + protégées)
		stats := v1.Group("/stats")
		{
			stats.GET("/market", statsH.Market)   // prix moyen par ville
			stats.GET("/popular", statsH.Popular) // biens les plus vus

			statsAuth := stats.Group("/")
			statsAuth.Use(middleware.JWTAuth(), middleware.RequireRole("agent", "admin"))
			{
				statsAuth.GET("/dashboard", statsH.Dashboard) // KPIs complets
			}
		}
	}

	return r
}
