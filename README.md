shstat - simple shell statistics
================================

These are some tools I often write one-liners for when looking at
data on the console. It's the same one-liners every time, so 
I may as well package them and save some repetition.

* fld extracts columns from input. Credit Mark-Jason Dominus for
the idea (I think Mark calls this "f").
* tally sums columns from input. "sum" was taken because it runs
a checksum on the linux distribution I use.
* hist produces a histogram from its input. This one is not entirely
trivial if you want to graph the output, which this can do.

Install
-------

`go get github.com/gaal/fld`

`go get github.com/gaal/tally`

`go get github.com/gaal/hist`

These tools are distributed under the MIT/X license.

Documentation
-------------

`godoc` [github.com/gaal/fld/fld](http://godoc.org/github.com/gaal/fld/fld)

`godoc` [github.com/gaal/tally/tally](http://godoc.org/github.com/gaal/tally/tally)

`godoc` [github.com/gaal/hist/hist](http://godoc.org/github.com/gaal/hist/hist)

Beta
----

I want to keep the interface as simple as possible. But if there's
a strong reason to change the behavior of one of these tools,
write me before the end of May 2016 and I'll consider it. After that
I'll be more conservative about breaking things.

If you want different output order from hist, make sure you know about
the -n and -k options to sort(1).

Contact
-------

Gaal Yahas, gaal@forum2.org.
