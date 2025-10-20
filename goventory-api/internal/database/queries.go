package database

import (
	"database/sql"
)

// insert a new host into the database.
func CreateHost(db *sql.DB, host *Host) (*Host, error) {
	stmt, err := db.Prepare("INSERT INTO hosts(hostname, ip_address, host_group) VALUES(?, ?, ?)")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	result, err := stmt.Exec(host.Hostname, host.IPAddress, host.HostGroup)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	host.ID = id
	return host, nil
}
