module github.com/nzenitram/relay-esp

go 1.20

require (
	github.com/gorilla/mux v1.8.1
	github.com/lib/pq v1.10.9
)

require (
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/joho/godotenv v1.5.1
	github.com/sendgrid/sendgrid-go v3.16.0+incompatible
	golang.org/x/crypto v0.27.0
	golang.org/x/exp v0.0.0-20230224173230-c95f2b4c22f2 // Use a version known to work with Go 1.20
)

require (
	github.com/sendgrid/rest v2.6.9+incompatible // indirect
	github.com/stretchr/testify v1.9.0 // indirect
)
