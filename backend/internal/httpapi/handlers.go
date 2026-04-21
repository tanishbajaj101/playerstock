package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stakestock/backend/internal/accounts"
	"github.com/stakestock/backend/internal/auth"
	"github.com/stakestock/backend/internal/models"
	"github.com/stakestock/backend/internal/trading"
)

// ---- Auth handlers ----

func (h *Handler) googleLogin(w http.ResponseWriter, r *http.Request) {
	state, err := auth.GenerateState()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errResp("failed to generate state"))
		return
	}
	auth.SetStateCookie(w, state, h.sessions.IsSecure())
	http.Redirect(w, r, h.google.AuthCodeURL(state), http.StatusTemporaryRedirect)
}

func (h *Handler) googleCallback(w http.ResponseWriter, r *http.Request) {
	expectedState := auth.GetStateCookie(r)
	if expectedState == "" || r.URL.Query().Get("state") != expectedState {
		writeJSON(w, http.StatusBadRequest, errResp("invalid state"))
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		writeJSON(w, http.StatusBadRequest, errResp("missing code"))
		return
	}

	info, err := h.google.Exchange(r.Context(), code)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errResp("oauth exchange failed"))
		return
	}

	user, err := h.sessions.UpsertUser(r.Context(), info.Sub, info.Email)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errResp("failed to upsert user"))
		return
	}

	if err := h.sessions.CreateSession(r.Context(), w, user.ID); err != nil {
		writeJSON(w, http.StatusInternalServerError, errResp("failed to create session"))
		return
	}

	frontendURL := h.frontendOrigin
	if user.Username == nil {
		frontendURL += "/onboarding"
	} else {
		frontendURL += "/"
	}
	http.Redirect(w, r, frontendURL, http.StatusTemporaryRedirect)
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	h.sessions.DeleteSession(r.Context(), w, r)
	writeJSON(w, http.StatusOK, map[string]string{"status": "logged out"})
}

// ---- User handlers ----

func (h *Handler) getMe(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	bal, err := accounts.GetBalance(r.Context(), h.pool, user.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errResp("failed to get balance"))
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"user":             user,
		"balance":          bal,
		"needs_onboarding": user.Username == nil,
	})
}

func (h *Handler) setUsername(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	var body struct {
		Username string `json:"username"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Username == "" {
		writeJSON(w, http.StatusBadRequest, errResp("username required"))
		return
	}

	_, err := h.pool.Exec(r.Context(), `UPDATE users SET username=$2 WHERE id=$1 AND username IS NULL`, user.ID, body.Username)
	if err != nil {
		writeJSON(w, http.StatusConflict, errResp("username taken"))
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"username": body.Username})
}

// ---- Asset handlers ----

func (h *Handler) listAssets(w http.ResponseWriter, r *http.Request) {
	rows, err := h.pool.Query(r.Context(), `
		SELECT a.id, a.symbol, a.name, a.description, a.nationality, a.role,
		       (SELECT t.price FROM trades t WHERE t.asset_id = a.id ORDER BY t.created_at DESC LIMIT 1) AS last_price,
		       (SELECT ps.price FROM price_snapshots ps WHERE ps.asset_id = a.id AND ps.ts <= now() - interval '24 hours' ORDER BY ps.ts DESC LIMIT 1) AS price_24h_ago,
		       COALESCE((SELECT SUM(t.qty) FROM trades t WHERE t.asset_id = a.id AND t.created_at > now() - interval '24 hours'), 0) AS volume_24h
		FROM assets a
		ORDER BY a.symbol
	`)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errResp("db error"))
		return
	}
	defer rows.Close()

	var assets []models.AssetWithPrice
	for rows.Next() {
		var a models.AssetWithPrice
		if err := rows.Scan(&a.ID, &a.Symbol, &a.Name, &a.Description, &a.Nationality, &a.Role, &a.LastPrice, &a.Price24hAgo, &a.Volume24h); err != nil {
			continue
		}
		if a.LastPrice != nil && a.Price24hAgo != nil && !a.Price24hAgo.IsZero() {
			pct := a.LastPrice.Sub(*a.Price24hAgo).Div(*a.Price24hAgo).Mul(decimal.NewFromInt(100))
			a.ChangePct = &pct
		}
		assets = append(assets, a)
	}
	if assets == nil {
		assets = []models.AssetWithPrice{}
	}
	writeJSON(w, http.StatusOK, assets)
}

func (h *Handler) getAsset(w http.ResponseWriter, r *http.Request) {
	symbol := chi.URLParam(r, "symbol")
	var a models.AssetWithPrice
	err := h.pool.QueryRow(r.Context(), `
		SELECT a.id, a.symbol, a.name, a.description, a.nationality, a.role,
		       (SELECT t.price FROM trades t WHERE t.asset_id = a.id ORDER BY t.created_at DESC LIMIT 1) AS last_price,
		       (SELECT ps.price FROM price_snapshots ps WHERE ps.asset_id = a.id AND ps.ts <= now() - interval '24 hours' ORDER BY ps.ts DESC LIMIT 1) AS price_24h_ago,
		       COALESCE((SELECT SUM(t.qty) FROM trades t WHERE t.asset_id = a.id AND t.created_at > now() - interval '24 hours'), 0) AS volume_24h
		FROM assets a WHERE a.symbol=$1
	`, symbol).Scan(&a.ID, &a.Symbol, &a.Name, &a.Description, &a.Nationality, &a.Role, &a.LastPrice, &a.Price24hAgo, &a.Volume24h)
	if err != nil {
		writeJSON(w, http.StatusNotFound, errResp("asset not found"))
		return
	}
	if a.LastPrice != nil && a.Price24hAgo != nil && !a.Price24hAgo.IsZero() {
		pct := a.LastPrice.Sub(*a.Price24hAgo).Div(*a.Price24hAgo).Mul(decimal.NewFromInt(100))
		a.ChangePct = &pct
	}
	writeJSON(w, http.StatusOK, a)
}

func (h *Handler) getDepth(w http.ResponseWriter, r *http.Request) {
	symbol := chi.URLParam(r, "symbol")
	ab, ok := h.registry.Get(symbol)
	if !ok {
		writeJSON(w, http.StatusNotFound, errResp("asset not found"))
		return
	}
	reply := ab.Depth()
	var bids, asks []models.PriceLevel
	for _, l := range reply.Bids {
		bids = append(bids, models.PriceLevel{Price: l.Price, Quantity: l.Quantity})
	}
	for _, l := range reply.Asks {
		asks = append(asks, models.PriceLevel{Price: l.Price, Quantity: l.Quantity})
	}
	if bids == nil {
		bids = []models.PriceLevel{}
	}
	if asks == nil {
		asks = []models.PriceLevel{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"bids": bids, "asks": asks})
}

func (h *Handler) getAssetTrades(w http.ResponseWriter, r *http.Request) {
	symbol := chi.URLParam(r, "symbol")
	rows, err := h.pool.Query(r.Context(), `
		SELECT t.id, t.asset_id, t.buy_order_id, t.sell_order_id,
		       t.buy_user_id, t.sell_user_id, t.qty, t.price, t.created_at
		FROM trades t
		JOIN assets a ON a.id = t.asset_id
		WHERE a.symbol=$1
		ORDER BY t.created_at DESC
		LIMIT 50
	`, symbol)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errResp("db error"))
		return
	}
	defer rows.Close()

	var trades []models.Trade
	for rows.Next() {
		var t models.Trade
		if err := rows.Scan(&t.ID, &t.AssetID, &t.BuyOrderID, &t.SellOrderID, &t.BuyUserID, &t.SellUserID, &t.Qty, &t.Price, &t.CreatedAt); err != nil {
			continue
		}
		trades = append(trades, t)
	}
	if trades == nil {
		trades = []models.Trade{}
	}
	writeJSON(w, http.StatusOK, trades)
}

// ---- Order handlers ----

func (h *Handler) placeOrder(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())

	var req trading.PlaceOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errResp("invalid request body"))
		return
	}

	order, err := h.trading.PlaceOrder(r.Context(), user.ID, req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errResp(err.Error()))
		return
	}
	writeJSON(w, http.StatusCreated, order)
}

func (h *Handler) cancelOrder(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	idStr := chi.URLParam(r, "id")
	orderID, err := uuid.Parse(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errResp("invalid order id"))
		return
	}
	if err := h.trading.CancelOrder(r.Context(), user.ID, orderID); err != nil {
		writeJSON(w, http.StatusBadRequest, errResp(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "cancelled"})
}

func (h *Handler) listOrders(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	statusFilter := r.URL.Query().Get("status")

	query := `
		SELECT o.id, o.user_id, o.asset_id, o.side, o.type, o.qty, o.filled_qty,
		       o.price, o.status, o.is_short, o.created_at, o.updated_at
		FROM orders o
		WHERE o.user_id = $1
	`
	args := []any{user.ID}
	if statusFilter != "" {
		query += fmt.Sprintf(" AND o.status = $%d", len(args)+1)
		args = append(args, statusFilter)
	}
	query += " ORDER BY o.created_at DESC LIMIT 100"

	rows, err := h.pool.Query(r.Context(), query, args...)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errResp("db error"))
		return
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var o models.Order
		var price *decimal.Decimal
		if err := rows.Scan(&o.ID, &o.UserID, &o.AssetID, &o.Side, &o.Type, &o.Qty, &o.FilledQty, &price, &o.Status, &o.IsShort, &o.CreatedAt, &o.UpdatedAt); err != nil {
			continue
		}
		o.Price = price
		orders = append(orders, o)
	}
	if orders == nil {
		orders = []models.Order{}
	}
	writeJSON(w, http.StatusOK, orders)
}

func (h *Handler) getPortfolio(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())

	bal, err := accounts.GetBalance(r.Context(), h.pool, user.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errResp("balance error"))
		return
	}

	rows, err := h.pool.Query(r.Context(), `
		SELECT p.user_id, p.asset_id, p.qty, p.locked_qty,
		       a.id, a.symbol, a.name, a.description, a.nationality, a.role,
		       (SELECT t.price FROM trades t WHERE t.asset_id = a.id ORDER BY t.created_at DESC LIMIT 1) AS last_price
		FROM positions p
		JOIN assets a ON a.id = p.asset_id
		WHERE p.user_id = $1 AND (p.qty != 0 OR p.locked_qty != 0)
	`, user.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errResp("db error"))
		return
	}
	defer rows.Close()

	var positions []models.PortfolioPosition
	for rows.Next() {
		var pp models.PortfolioPosition
		var lastPrice *decimal.Decimal
		err := rows.Scan(
			&pp.UserID, &pp.AssetID, &pp.Qty, &pp.LockedQty,
			&pp.Asset.ID, &pp.Asset.Symbol, &pp.Asset.Name, &pp.Asset.Description, &pp.Asset.Nationality, &pp.Asset.Role,
			&lastPrice,
		)
		if err != nil {
			continue
		}
		pp.LastPrice = lastPrice
		if lastPrice != nil {
			pnl := pp.Qty.Mul(*lastPrice)
			pp.UnrealPnL = &pnl
		}
		positions = append(positions, pp)
	}
	if positions == nil {
		positions = []models.PortfolioPosition{}
	}

	// Open orders
	orows, err := h.pool.Query(r.Context(), `
		SELECT o.id, o.user_id, o.asset_id, o.side, o.type, o.qty, o.filled_qty,
		       o.price, o.status, o.is_short, o.created_at, o.updated_at
		FROM orders o
		WHERE o.user_id=$1 AND o.status IN ('open','partial')
		ORDER BY o.created_at DESC
	`, user.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errResp("db error"))
		return
	}
	defer orows.Close()

	var openOrders []models.Order
	for orows.Next() {
		var o models.Order
		var price *decimal.Decimal
		if err := orows.Scan(&o.ID, &o.UserID, &o.AssetID, &o.Side, &o.Type, &o.Qty, &o.FilledQty, &price, &o.Status, &o.IsShort, &o.CreatedAt, &o.UpdatedAt); err != nil {
			continue
		}
		o.Price = price
		openOrders = append(openOrders, o)
	}
	if openOrders == nil {
		openOrders = []models.Order{}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"balance":     bal,
		"positions":   positions,
		"open_orders": openOrders,
	})
}

func (h *Handler) getHistory(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	rows, err := h.pool.Query(r.Context(), `
		SELECT t.id, t.asset_id, t.buy_order_id, t.sell_order_id,
		       t.buy_user_id, t.sell_user_id, t.qty, t.price, t.created_at
		FROM trades t
		WHERE t.buy_user_id=$1 OR t.sell_user_id=$1
		ORDER BY t.created_at DESC
		LIMIT 100
	`, user.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errResp("db error"))
		return
	}
	defer rows.Close()

	var trades []models.Trade
	for rows.Next() {
		var t models.Trade
		if err := rows.Scan(&t.ID, &t.AssetID, &t.BuyOrderID, &t.SellOrderID, &t.BuyUserID, &t.SellUserID, &t.Qty, &t.Price, &t.CreatedAt); err != nil {
			continue
		}
		trades = append(trades, t)
	}
	if trades == nil {
		trades = []models.Trade{}
	}
	writeJSON(w, http.StatusOK, trades)
}

// ---- Chart handler ----

func (h *Handler) getAssetChart(w http.ResponseWriter, r *http.Request) {
	symbol := chi.URLParam(r, "symbol")

	var stepSecs, lookbackSecs int
	switch r.URL.Query().Get("tf") {
	case "7d":
		stepSecs, lookbackSecs = 4*3600, 7*24*3600
	case "30d":
		stepSecs, lookbackSecs = 16*3600, 30*24*3600
	default: // "24h"
		stepSecs, lookbackSecs = 1800, 24*3600
	}

	var assetID string
	if err := h.pool.QueryRow(r.Context(),
		`SELECT id FROM assets WHERE symbol=$1`, symbol,
	).Scan(&assetID); err != nil {
		writeJSON(w, http.StatusNotFound, errResp("asset not found"))
		return
	}

	rows, err := h.pool.Query(r.Context(), `
		SELECT
			floor(extract(epoch from ts) / $2)::bigint * $2 * 1000 AS ts_ms,
			avg(price)::float8 AS price
		FROM price_snapshots
		WHERE asset_id = $1
		  AND ts >= now() - ($3 * interval '1 second')
		GROUP BY floor(extract(epoch from ts) / $2)::bigint
		ORDER BY 1
	`, assetID, stepSecs, lookbackSecs)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errResp("db error"))
		return
	}
	defer rows.Close()

	points := []models.PricePoint{}
	for rows.Next() {
		var p models.PricePoint
		if err := rows.Scan(&p.Ts, &p.Price); err == nil {
			points = append(points, p)
		}
	}
	writeJSON(w, http.StatusOK, points)
}

func (h *Handler) getCharts(w http.ResponseWriter, r *http.Request) {
	var stepSecs, lookbackSecs int
	switch r.URL.Query().Get("tf") {
	case "7d":
		stepSecs, lookbackSecs = 4*3600, 7*24*3600
	case "30d":
		stepSecs, lookbackSecs = 16*3600, 30*24*3600
	default: // "24h"
		stepSecs, lookbackSecs = 1800, 24*3600
	}

	rows, err := h.pool.Query(r.Context(), `
		SELECT
			a.symbol,
			floor(extract(epoch from ps.ts) / $1)::bigint * $1 * 1000 AS ts_ms,
			avg(ps.price)::float8 AS price
		FROM assets a
		JOIN price_snapshots ps ON ps.asset_id = a.id
		WHERE ps.ts >= now() - ($2 * interval '1 second')
		GROUP BY a.symbol, floor(extract(epoch from ps.ts) / $1)::bigint
		ORDER BY a.symbol, 2
	`, stepSecs, lookbackSecs)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errResp("db error"))
		return
	}
	defer rows.Close()

	out := map[string][]models.PricePoint{}
	for rows.Next() {
		var symbol string
		var p models.PricePoint
		if err := rows.Scan(&symbol, &p.Ts, &p.Price); err == nil {
			out[symbol] = append(out[symbol], p)
		}
	}
	writeJSON(w, http.StatusOK, out)
}

// ---- Helpers ----

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func errResp(msg string) map[string]string {
	return map[string]string{"error": msg}
}

