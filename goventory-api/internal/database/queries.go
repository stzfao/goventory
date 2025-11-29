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

// retrieve a single host by its hostname.
func GetHostByHostname(db *sql.DB, hostname string) (*Host, error) {
	row := db.QueryRow("SELECT id, hostname, ip_address, host_group FROM hosts WHERE hostname = ?", hostname)

	var host Host
	err := row.Scan(&host.ID, &host.Hostname, &host.IPAddress, &host.HostGroup)
	if err != nil {
		return nil, err
	}

	return &host, nil
}

// removes a host from the database by its hostname.
func DeleteHostByHostname(db *sql.DB, hostname string) error {
	res, err := db.Exec("DELETE FROM hosts WHERE hostname = ?", hostname)
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
	res, err := db.Exec("UPDATE hosts SET ip_address = ?, host_group = ? WHERE hostname = ?", host.IPAddress, host.HostGroup, host.Hostname)
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
