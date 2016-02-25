# wordstat

$ echo "go bla bla-bla bla foo foo foo bar boo" | nc localhost 9000

$ curl http://localhost:8000/?N=3

{"top_words":["foo", "bla", "bar"]}
