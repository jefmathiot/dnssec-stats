# DNSSEC Stats

Extract DNSSEC support statistics for the top 1 million hosts of the
[Alexa](http://alexa.com) database.

The tool wraps `dig` to check, for each host present in an input file
(`top-1m.csv`), whether an `RRSIG` record is present in the command output.

You can grab a fresh copy of the list of the top one million websites
[here](http://s3.amazonaws.com/alexa-static/top-1m.csv.zip). At the time of
speaking (September 2016), 4.6194% of the hosts have an `RRSIG` record.

## Installation

Install the [Go release](https://golang.org/dl/) and clone this repository.

## Usage

The tool expects a `top-1m.csv` file to be present in the current directory.

Launch using:

```
go run dnssec-stats.go
```

The detailed results are saved to a `results.csv` file.

`dnssec-stats` defaults to 100 concurrent workers, each of them spawning `dig`
processes for a couple of hours, thus the execution is quite CPU-intensive. Feel
free to adjust the number of workers in the `main` function if you want to
reduce the load (the overall execution time will increase).

## License

Do What The Fuck You Want To Public License, Version 2, as published by Sam
Hocevar. See the LICENSE.txt file for details.
