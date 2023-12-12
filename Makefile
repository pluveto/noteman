default:
	go build cmd/catdate/catdate.go
	go build cmd/noteman/noteman.go

install:
	go install cmd/catdate/catdate.go
	go install cmd/noteman/noteman.go
