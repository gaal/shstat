shstat - simple shell statistics
================================

These are some tools I often write one-liners for when looking at
data on the console. It's the same one-liners every time, so 
I may as well package them and save some repetition.

* `fld` extracts columns from input. Credit Mark-Jason Dominus for
  the idea (I think Mark calls this "f").

        ps aux | grep myproc | fld 2

* `uni` identifies and enumerates Unicode characters. Credit Larry Wall
  for the idea and original implementation.

        uni 𝄞
        uni math fraktur
        uni 200a

* `tally` sums columns from input. "sum" was taken because it runs
  a checksum on the linux distribution I use.

        # sum file sizes
        ls -l | tally 5

* `hist` produces a histogram from its input. This one is not entirely
  trivial if you want to graph the output, which hist does.

        # The distribution of words in Hamlet follows a power law.
        # Default graph is linear, can use -scale=log or -graph=false to turn off.
        wget -O - https://www.gutenberg.org/cache/epub/1524/pg1524.txt | \
            grep -A9999 HAMLET | hist -words


Install
-------

    go get github.com/gaal/shstat/fld
    go get github.com/gaal/shstat/hist
    go get github.com/gaal/shstat/tally
    go get github.com/gaal/shstat/ucd
    go get github.com/gaal/shstat/uni

These tools are distributed under the MIT/X license.

Documentation
-------------

`godoc` [github.com/gaal/shstat/fld](http://godoc.org/github.com/gaal/shstat/fld)  
`godoc` [github.com/gaal/shstat/hist](http://godoc.org/github.com/gaal/shstat/hist) 
`godoc` [github.com/gaal/shstat/tally](http://godoc.org/github.com/gaal/shstat/tally)  
`godoc` [github.com/gaal/shstat/ucd](http://godoc.org/github.com/gaal/shstat/ucd)  
`godoc` [github.com/gaal/shstat/uni](http://godoc.org/github.com/gaal/shstat/uni)  

Beta
----

I want to keep the interface as simple as possible. But if there's
a strong reason to change the behavior of one of these tools,
write me before the end of May 2016 and I'll consider it. After that
I'll be more conservative about breaking things.

If you want different output order from hist, make sure you know about
the `-n` and `-k` options to `sort(1)`.

Contact
-------

Gaal Yahas, <gaal@forum2.org>.
