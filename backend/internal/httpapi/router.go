package httpapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/stakestock/backend/internal/auth"
	"github.com/stakestock/backend/internal/engine"
	"github.com/stakestock/backend/internal/pubsub"
	"github.com/stakestock/backend/internal/trading"
	"github.com/stakestock/backend/internal/ws"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	pool     *pgxpool.Pool
	google   *auth.GoogleProvider
	sessions *auth.SessionStore
	registry *engine.Registry
	trading  *trading.Service
	bus      *pubsub.Bus
	hub      *ws.Hub
	frontendOrigin string
}

func NewHandler(
	pool *pgxpool.Pool,
	google *auth.GoogleProvider,
	sessions *auth.SessionStore,
	registry *engine.Registry,
	tradingSvc *trading.Service,
	bus *pubsub.Bus,
	hub *ws.Hub,
	frontendOrigin string,
) *Handler {
	return &Handler{
		pool:           pool,
		google:         google,
		sessions:       sessions,
		registry:       registry,
		trading:        tradingSvc,
		bus:            bus,
		hub:            hub,
		frontendOrigin: frontendOrigin,
	}
}

func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(h.corsMiddleware)

	// Auth routes
	r.Get("/auth/google/login", h.googleLogin)
	r.Get("/auth/google/callback", h.googleCallback)
	r.Post("/auth/logout", h.logout)

	// WebSocket
	r.Get("/ws", h.hub.ServeHTTP)

	// API routes
	r.Route("/api", func(r chi.Router) {
		r.Use(h.requireAuth)

		r.Get("/me", h.getMe)
		r.Post("/me/username", h.setUsername)

		r.Get("/assets", h.listAssets)
		r.Get("/assets/{symbol}", h.getAsset)
		r.Get("/assets/{symbol}/depth", h.getDepth)
		r.Get("/assets/{symbol}/trades", h.getAssetTrades)
		r.Get("/assets/{symbol}/chart", h.getAssetChart)
		r.Get("/charts", h.getCharts)

		r.With(h.requireOnboarded).Post("/orders", h.placeOrder)
		r.With(h.requireOnboarded).Delete("/orders/{id}", h.cancelOrder)
		r.Get("/orders", h.listOrders)
		r.Get("/portfolio", h.getPortfolio)
		r.Get("/history", h.getHistory)
	})

	return r
}

func (h *Handler) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", h.frontendOrigin)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (h *Handler) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := h.sessions.GetUserFromRequest(r)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		next.ServeHTTP(w, r.WithContext(auth.WithUser(r.Context(), user)))
	})
}

func (h *Handler) requireOnboarded(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := auth.UserFromContext(r.Context())
		if user == nil || user.Username == nil {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "username required"})
			return
		}
		next.ServeHTTP(w, r)
	})
}
