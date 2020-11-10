test: ;go clean -testcache; go test  ./... --cover;

.PHONY:  test