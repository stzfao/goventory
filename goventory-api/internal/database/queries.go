package database

import (
	"database/sql"

	"github.com/google/uuid"
)

// CreateHost inserts a new host into the database with a UUID.
func CreateHost(db *sql.DB, host *Host) (*Host, error) {
	// Generate a new UUID for the host.
	host.ID = uuid.New().String()

	stmt, err := db.Prepare("INSERT INTO hosts(id, hostname, ip_address, host_group) VALUES(?, ?, ?, ?)")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	// Execute the statement with all four values.
	_, err = stmt.Exec(host.ID, host.Hostname, host.IPAddress, host.HostGroup)
	if err != nil {
		return nil, err
	}

	// Since we generated the ID, we just return the host object.
	return host, nil
}

// ListHosts retrieves all hosts from the database.
func ListHosts(db *sql.DB) ([]Host, error) {
	rows, err := db.Query("SELECT id, hostname, ip_address, host_group FROM hosts")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hosts []Host
	for rows.Next() {
		var host Host
		if err := rows.Scan(&host.ID, &host.Hostname, &host.IPAddress, &host.HostGroup); err != nil {
			return nil, err
		}
		hosts = append(hosts, host)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	if hosts == nil{
		return []Host{}, err
	}

	return hosts, nil
}

// GetHostByID retrieves a single host by its ID.
func GetHostByID(db *sql.DB, id string) (*Host, error) { // Changed id to string
	row := db.QueryRow("SELECT id, hostname, ip_address, host_group FROM hosts WHERE id = ?", id)

	var host Host
	err := row.Scan(&host.ID, &host.Hostname, &host.IPAddress, &host.HostGroup)
	if err != nil {
		return nil, err
	}

	return &host, nil
}

// DeleteHostByID removes a host from the database by its ID.
func DeleteHostByID(db *sql.DB, id string) error { // Changed id to string
	res, err := db.Exec("DELETE FROM hosts WHERE id = ?", id)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// UpdateHost modifies an existing host in the database.
func UpdateHost(db *sql.DB, host *Host) error {
	res, err := db.Exec("UPDATE hosts SET hostname = ?, ip_address = ?, host_group = ? WHERE id = ?", host.Hostname, host.IPAddress, host.HostGroup, host.ID)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}
