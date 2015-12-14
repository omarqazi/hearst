all : test bin 
	@echo Built all
bin : 
	go build
test :
	go test . ./auth ./datastore ./controller
run : bin
	./hearst
testrun : test run
	@echo Tested and Ran Hearst
clean :
	rm -rf hearst