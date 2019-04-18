const conf  = `
settle = "1s"
poll = "200ms"

# Database path to save state between shutdowns
database = "/Users/dhorsley/fwatch.db"


[files.catalogs]
paths = [
    "/Users/dhorsley/test/test.txt",
]

[files.document]
recursive = true
paths = [ "/Users/dhorsley/Documents"]
`
