package database

// PanelDomainPasskeyAccount exposes counts, never credential IDs or key material.
type PanelDomainPasskeyAccount struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Hostname string `json:"hostname"`
	Count    int    `json:"count"`
}

func (db *DB) PanelDomainPasskeyConflicts(hostname string) ([]PanelDomainPasskeyAccount, error) {
	rows, err := db.conn.Query(`SELECT u.id,u.username,p.rp_id,COUNT(*) FROM passkeys p JOIN users u ON u.id=p.user_id WHERE p.rp_id<>? GROUP BY u.id,u.username,p.rp_id ORDER BY u.username,p.rp_id`, hostname)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	accounts := make([]PanelDomainPasskeyAccount, 0)
	for rows.Next() {
		var account PanelDomainPasskeyAccount
		if err := rows.Scan(&account.UserID, &account.Username, &account.Hostname, &account.Count); err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}
	return accounts, rows.Err()
}
