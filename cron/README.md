Usage

Callers may register Funcs to be invoked on a given schedule.  Cron will run
them in their own goroutines.

```
    c := cron.New()
    c.AddFunc("every half hour", "30 * * * *", func() { fmt.Println("Every hour on the half hour") })
    c.Start()
    ..
    // Output can be set to c, which will output the info when entry is going to run.
    c.SetOutput(o)
    ...
    // Funcs may also be added to a running Cron
    c.AddFunc("daily work", "0 1 * * *", func() { fmt.Println("Every day") })
    ..
    c.Stop()  // Stop the scheduler (does not stop any jobs already running).
```

CRON Expression Format

A cron expression represents a set of times, using 6 space-separated fields.

    Field name   | Mandatory? | Allowed values  | Allowed special characters
    ----------   | ---------- | --------------  | --------------------------
    Minutes      | Yes        | 0-59            | * / , -
    Hours        | Yes        | 0-23            | * / , -
    Day of month | Yes        | 1-31            | * / , -
    Month        | Yes        | 1-12            | * / , -
    Day of week  | Yes        | 0-6             | * / , -


Special Characters

Asterisk ( * )

The asterisk indicates that the cron expression will match for all values of the
field; e.g., using an asterisk in the 5th field (month) would indicate every
month.

Slash ( / )

Slashes are used to describe increments of ranges. For example 3-59/15 in the
1st field (minutes) would indicate the 3rd minute of the hour and every 15
minutes thereafter. The form "*\/..." is equivalent to the form "first-last/...",
that is, an increment over the largest possible range of the field.  The form
"N/..." is accepted as meaning "N-MAX/...", that is, starting at N, use the
increment until the end of that specific range.  It does not wrap around.

Comma ( , )

Commas are used to separate items of a list. For example, using "1,2,5" in
the 5th field (day of week) would mean Mondays, Wednesdays and Fridays.

Hyphen ( - )

Hyphens are used to define ranges. For example, 9-17 would indicate every
hour between 9am and 5pm inclusive.

Question mark ( ? )

Question mark may be used instead of '*' for leaving either day-of-month or
day-of-week blank.

Thread safety

Since the Cron service runs concurrently with the calling code, some amount of
care must be taken to ensure proper synchronization.

All cron methods are designed to be correctly synchronized as long as the caller
ensures that invocations have a clear happens-before ordering between them.

Implementation

Cron entries are stored in an array, sorted by their next activation time.  Cron
sleeps until the next job is due to be run.

Upon waking:
 - it runs each entry that is active on that second
 - it calculates the next run times for the jobs that were run
 - it goes to sleep until the soonest job.
 