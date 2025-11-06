package orderrepo

import (
	pb "349877-artemkagor05-course-1478/gen/api"
	"context"
	"errors"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lib/pq"
)

var (
	ErrOrderNotFound      = errors.New("order not found")
	ErrOrderAlreadyExists = errors.New("order already exists")
)

type OrderRepository struct {
	db      *pgxpool.Pool
	builder squirrel.StatementBuilderType
}

func New(db *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{
		db:      db,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *OrderRepository) CreateOrder(ctx context.Context, item string, quantity int32) (string, error) {
	query, args, err := r.builder.Insert("orders").
		Columns("item", "quantity").
		Values(item, quantity).
		Suffix("RETURNING id").
		ToSql()
	if err != nil {
		return "", fmt.Errorf("failed to build query: %w", err)
	}

	var newOrderID string
	if err = r.db.QueryRow(ctx, query, args...).Scan(&newOrderID); err != nil {
		if errors.Is(err, &pq.Error{Code: "23505"}) {
			return "", ErrOrderAlreadyExists
		}
		return "", fmt.Errorf("failed to execute query: %w", err)
	}

	return newOrderID, nil
}

func (r *OrderRepository) GetOrder(ctx context.Context, id string) (*pb.Order, error) {
	query, args, err := r.builder.Select("id", "item", "quantity").
		From("orders").
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var order pb.Order
	if err = r.db.QueryRow(ctx, query, args...).Scan(&order.Id, &order.Item, &order.Quantity); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrOrderNotFound
		}
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	return &order, nil
}

func (r *OrderRepository) UpdateOrder(ctx context.Context, id string, item string, quantity int32) (*pb.Order, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			rbErr := tx.Rollback(ctx)
			if rbErr != nil {
				fmt.Printf("failed to rollback transaction: %v\n", rbErr)
			}
			return
		}
		if commitErr := tx.Commit(ctx); commitErr != nil {
			err = fmt.Errorf("failed to commit transaction: %w", commitErr)
		}
	}()

	queryCheckExists, argsCheckExists, err := r.builder.Select("id", "item", "quantity").
		From("orders").
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build exists query: %w", err)
	}

	var currentOrder pb.Order
	if err = tx.QueryRow(ctx, queryCheckExists, argsCheckExists...).
		Scan(&currentOrder.Id, &currentOrder.Item, &currentOrder.Quantity); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrOrderNotFound
		}
		return nil, fmt.Errorf("failed to execute query (check existence): %w", err)
	}

	queryCheckUnique, argsCheckUnique, err := r.builder.Select("id").
		From("orders").
		Where(squirrel.And{
			squirrel.Eq{"item": item},
			squirrel.NotEq{"id": id},
		}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build uniqueness query: %w", err)
	}

	var existingID string
	if err = tx.QueryRow(ctx, queryCheckUnique, argsCheckUnique...).Scan(&existingID); err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("failed to check uniqueness: %w", err)
		} // else ok
	} else {
		return nil, ErrOrderAlreadyExists
	}

	queryUpdate, argsUpdate, err := r.builder.Update("orders").
		Set("item", item).
		Set("quantity", quantity).
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build update query: %w", err)
	}

	result, err := tx.Exec(ctx, queryUpdate, argsUpdate...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute update query: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected != 1 {
		return nil, ErrOrderNotFound
	}

	updatedOrder := &pb.Order{
		Id:       id,
		Item:     item,
		Quantity: quantity,
	}
	err = nil

	return updatedOrder, nil
}

func (r *OrderRepository) DeleteOrder(ctx context.Context, id string) error {
	query, args, err := r.builder.Delete("orders").
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	result, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrOrderNotFound
	}

	return nil
}

func (r *OrderRepository) ListOrders(ctx context.Context) []*pb.Order {
	query, args, err := r.builder.Select("id", "item", "quantity").
		From("orders").
		ToSql()
	if err != nil {
		fmt.Printf("failed to build query: %e", err)
		return nil
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		fmt.Printf("failed to execute query: %e", err)
		return nil
	}
	defer rows.Close()

	var orders []*pb.Order
	for rows.Next() {
		var order pb.Order
		if err = rows.Scan(&order.Id, &order.Item, &order.Quantity); err != nil {
			fmt.Printf("failed to scan rows: %e", err)
			return nil
		}
		orders = append(orders, &order)
	}

	if err = rows.Err(); err != nil {
		fmt.Printf("row error: %e", err)
		return nil
	}

	return orders
}
