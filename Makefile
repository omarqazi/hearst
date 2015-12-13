all : test bin 
	@echo Built all
bin : 
	go build
test :
	go test -v . ./datastore