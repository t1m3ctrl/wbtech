package repository

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"wbtech"
)

type OrderPostgres struct {
	db *sqlx.DB
}

func NewOrderPostgres(db *sqlx.DB) *OrderPostgres {
	return &OrderPostgres{db: db}
}

func (r *OrderPostgres) GetOrderById(_ context.Context, id string) (wbtech.Order, error) {
	var order wbtech.Order

	orderQuery := `SELECT * FROM orders WHERE order_uid = $1`
	err := r.db.Get(&order, orderQuery, id)
	if err != nil {
		return wbtech.Order{}, fmt.Errorf("failed to get order: %w", err)
	}

	deliveryQuery := `SELECT * FROM deliveries WHERE order_uid = $1`
	err = r.db.Get(&order.Delivery, deliveryQuery, id)
	if err != nil {
		return wbtech.Order{}, fmt.Errorf("failed to get delivery: %w", err)
	}

	paymentQuery := `SELECT * FROM payments WHERE order_uid = $1`
	err = r.db.Get(&order.Payment, paymentQuery, id)
	if err != nil {
		return wbtech.Order{}, fmt.Errorf("failed to get payment: %w", err)
	}

	itemsQuery := `SELECT * FROM items WHERE order_uid = $1`
	err = r.db.Select(&order.Items, itemsQuery, id)
	if err != nil {
		return wbtech.Order{}, fmt.Errorf("failed to get items: %w", err)
	}

	return order, nil
}

func (r *OrderPostgres) CreateOrder(ctx context.Context, order wbtech.Order) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func(tx *sqlx.Tx) {
		err := tx.Rollback()
		if err != nil {

		}
	}(tx)

	orderQuery := `INSERT INTO orders (
		order_uid, track_number, entry, locale, internal_signature,
		customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err = tx.ExecContext(ctx, orderQuery,
		order.OrderUID,
		order.TrackNumber,
		order.Entry,
		order.Locale,
		order.InternalSignature,
		order.CustomerID,
		order.DeliveryService,
		order.Shardkey,
		order.SmID,
		order.DateCreated,
		order.OofShard,
	)

	if err != nil {
		return fmt.Errorf("failed to insert order: %w", err)
	}

	deliveryQuery := `INSERT INTO deliveries (
		order_uid, name, phone, zip, city, address, region, email
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err = tx.ExecContext(ctx, deliveryQuery,
		order.OrderUID,
		order.Delivery.Name,
		order.Delivery.Phone,
		order.Delivery.Zip,
		order.Delivery.City,
		order.Delivery.Address,
		order.Delivery.Region,
		order.Delivery.Email,
	)

	if err != nil {
		return fmt.Errorf("failed to insert delivery: %w", err)
	}

	paymentQuery := `INSERT INTO payments (
		order_uid, transaction, request_id, currency, provider,
		amount, payment_dt, bank, delivery_cost, goods_total, custom_fee
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err = tx.ExecContext(ctx, paymentQuery,
		order.OrderUID,
		order.Payment.Transaction,
		order.Payment.RequestID,
		order.Payment.Currency,
		order.Payment.Provider,
		order.Payment.Amount,
		order.Payment.PaymentDt,
		order.Payment.Bank,
		order.Payment.DeliveryCost,
		order.Payment.GoodsTotal,
		order.Payment.CustomFee,
	)

	if err != nil {
		return fmt.Errorf("failed to insert payment: %w", err)
	}

	itemQuery := `INSERT INTO items (
		order_uid, chrt_id, track_number, price, rid, name,
		sale, size, total_price, nm_id, brand, status
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	for _, item := range order.Items {
		_, err = tx.ExecContext(ctx, itemQuery,
			order.OrderUID,
			item.ChrtID,
			item.TrackNumber,
			item.Price,
			item.Rid,
			item.Name,
			item.Sale,
			item.Size,
			item.TotalPrice,
			item.NmID,
			item.Brand,
			item.Status,
		)

		if err != nil {
			return fmt.Errorf("failed to insert item: %w", err)
		}
	}

	return tx.Commit()
}
