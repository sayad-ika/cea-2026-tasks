package main

type User struct {
	ID                    int
	Name, Email, Password string
}

var users = map[string]User{
	"doha@craftsmensoftware.com":      {1, "Ali Haider Doha", "doha@craftsmensoftware.com", "password123"},
	"sayad.ibn@craftsmensoftware.com": {2, "Sayad Ibn K A", "sayad.ibn@craftsmensoftware.com", "letmein"},
}
