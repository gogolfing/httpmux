## Requirements

Following are the acceptance requirements for the performance of this package.

- Route paths must alternate between variable and static parts.
    - Creating sub routes assume a beginning path separator on the new path if
    the existing sub route does not end with one.
- Segment variable names read until the next `/` or the end of input including
`:` and `*`.
- End variable names read until the end of input including `/`, `:`, and `*`.
- Static parts and segment variables must not overlap.
- Static parts and end variables may overlap.
    - The static part is served if it can successfully be served like any other
    route starting with a static part. Otherwise, the end variable route is served.
- Segment and end variables must not overlap.
- Variable names (both static and end) may be empty.
