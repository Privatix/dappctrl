package data

import (
	"fmt"

	"gopkg.in/reform.v1"
)

// Recover ensures data consistency after unexpected controller's exit.
func Recover(db *reform.DB) error {
	if err := recoverServiceStatuses(db); err != nil {
		return err
	}

	return nil
}

// recoverServiceStatuses updates service_status for client channels:
// activating, active and suspending becomes suspended, while terminating
// becomes terminated.
func recoverServiceStatuses(db *reform.DB) error {
	format := `
		UPDATE channels
		   SET service_status = '%s'
		  FROM offerings, products
		 WHERE offering = offerings.id
		          AND product = products.id
		          AND NOT is_server
		          AND service_status %s`

	query := fmt.Sprintf(format, ServiceSuspended,
		"IN ('activating', 'active', 'suspending')")
	if _, err := db.Exec(query); err != nil {
		return err
	}

	query = fmt.Sprintf(format, ServiceTerminated, " = 'terminating'")
	if _, err := db.Exec(query); err != nil {
		return err
	}

	return nil
}
