package sqlx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"myfacebook-dialog/internal/db"
	"myfacebook-dialog/internal/repository"
)

type DialogRepository struct {
	db *db.DB
}

func NewDialogRepository(db *db.DB) *DialogRepository {
	return &DialogRepository{
		db: db,
	}
}

func (r *DialogRepository) Add(ctx context.Context, dialogMessage repository.DialogMessage) error {
	dbConn := r.db.GetConnection()

	sqlQuery := `INSERT INTO dialogs (sender_id, receiver_id, text) 
				VALUES (:sender_id, :receiver_id, :text)`

	_, err := dbConn.NamedExecContext(ctx, sqlQuery, dialogMessage)
	if err != nil {
		return fmt.Errorf("failed to add dialog mesage to db: %w", err)
	}

	return nil
}

func (r *DialogRepository) GetDialogMessagesBySenderIDAndReceiverID(ctx context.Context, senderID, receiverID string) ([]repository.DialogMessage, error) {
	dbConn := r.db.GetConnection()

	var dialogMessages []repository.DialogMessage

	sqlQuery := `SELECT id, sender_id, receiver_id, text 
		FROM dialogs WHERE (sender_id=$1 AND receiver_id=$2) OR (sender_id=$2 AND receiver_id=$1) 
		ORDER BY created_at`

	err := dbConn.SelectContext(ctx, &dialogMessages, sqlQuery, senderID, receiverID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrNotFound
		}

		return nil, fmt.Errorf("failed to fetch dialog messages by senderID and receiverID: %w", err)
	}

	return dialogMessages, nil
}
