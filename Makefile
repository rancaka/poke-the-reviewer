run:
	@echo "${NOW} == BUILDING PokeTheReviewer"
	@GOOS=linux go build -o poke-the-reviewer *.go