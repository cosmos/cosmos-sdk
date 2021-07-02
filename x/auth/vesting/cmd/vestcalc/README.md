# `vestcalc`: A vesting schedule calculator

A periodic vesting account has its vesting schedule configured as a sequence
of vesting events, spaced by the relative time between them, in seconds.
Most vesting agreements, however, are specified in terms of a number of
monthly events from a given start time, possibly subject to one or more
vesting "cliffs" which delay vesting until at or after the cliff.

This tool can generate a vesting schedule given the parameters above,
and can translate a vesting schedule into readable timestamps.

This tool correctly handles:

- clipping event dates to the end of the month (e.g. vesting on the 31st of
  the month happens on the 30th in June);
- daylight saving time;
- leap years;
- large amounts (up to 9 quintillion).

All times are interpreted in the local timezone, since the desired vesting
schedule is commonly specified in local time. To use another timezone,
set your `TZ` environment variable before running the command.

## Writing a schedule

When the `--write` flag is set, the tool will write a schedule in JSON to
stdout. The following flags control the output:

- `--amount`: A decimal number giving the total amount to vest.
- `--denom`: The denomination for the amount.  Current valid options are:
    - `ubld`: one millionth of an Agoric BLD
- `--months`: The number of months to vest over.
- `--time`: The time of day of the vesting event, in 24-hour HH:MM format.
  Defaults to midnight.
- `--start`: The vesting start time: i.e. the first event happens in the
  next month. Specified in the format `YYYY-MM-DD` or `YYYY-MM-DDThh:mm`,
  e.g. `2006-01-02T15:04` for 3:04pm on January 2, 2006.
- `--cliffs`: One or more vesting cliffs in `YYYY-MM-DD` or `YYYY-MM-DDThh:mm`
  format. Only the latest one will have any effect, but it is useful to let
  the computer do that calculation to avoid mistakes. Multiple cliff dates
  can be separated by commas or given as multiple arguments.

## Reading a schedule

When the `--read` flag is set, th tool will read a schedule in JSON from
stdin and write the vesting events in absolute time to stdout. The following
flags control the command:

- `--start`: The vesting start time used when generating the schedule.
  Specified in the format `YYYY-MM-DD` or `YYYY-MM-DDThh:mm`, e.g.
  `2006-01-02T15:04` for 3:04pm on January 2, 2006.

## Examples

```
$ vestcalc --write --start=2021-01-01 --amount=1000000000 --denom=ubld \
> --months=24 --time=09:00 --cliffs=2022-01-15T00:00 | \
> vestcalc --read --start=2021-01-01
[
    2022-01-15T00:00: 500000000
    2022-02-01T09:00: 41666666
    2022-03-01T09:00: 41666667
    2022-04-01T09:00: 41666667
    2022-05-01T09:00: 41666666
    2022-06-01T09:00: 41666667
    2022-07-01T09:00: 41666667
    2022-08-01T09:00: 41666666
    2022-09-01T09:00: 41666667
    2022-10-01T09:00: 41666667
    2022-11-01T09:00: 41666666
    2022-12-01T09:00: 41666667
    2023-01-01T09:00: 41666667
]
$ vestcalc --write --start=2021-01-01 --amount=1000000000 --denom=ubld \
> --months=24 --time=09:00 --cliffs=2022-01-15T00:00
[
  {
    "coins": "500000000ubld",
    "length_seconds": 32711400
  },
  {
    "coins": "41666666ubld",
    "length_seconds": 1501200
  },
  {
    "coins": "41666667ubld",
    "length_seconds": 2419200
  },
  {
    "coins": "41666667ubld",
    "length_seconds": 2674800
  },
  {
    "coins": "41666666ubld",
    "length_seconds": 2592000
  },
  {
    "coins": "41666667ubld",
    "length_seconds": 2678400
  },
  {
    "coins": "41666667ubld",
    "length_seconds": 2592000
  },
  {
    "coins": "41666666ubld",
    "length_seconds": 2678400
  },
  {
    "coins": "41666667ubld",
    "length_seconds": 2678400
  },
  {
    "coins": "41666667ubld",
    "length_seconds": 2592000
  },
  {
    "coins": "41666666ubld",
    "length_seconds": 2678400
  },
  {
    "coins": "41666667ubld",
    "length_seconds": 2595600
  },
  {
    "coins": "41666667ubld",
    "length_seconds": 2678400
  }
]
```
