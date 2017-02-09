## Requirements

Following are the acceptance requirements for the performance of this package.

#### Creating Routes
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

#### Finding and Serving Routes
- All incoming request paths are cleaned. See ./httpmux/path.Clean().
- All of the registered path must be matched in the request path.
    - If there is an ending path separator in the registered path, then the request
    path must also end with a path separator. There will never be multiple ending
    path separators due the aforementioned clean.

- When matching a segment variable, the value matches until the next path separator
or end of the request path.
- Segment variables starting immediately after a path separator must not have
an empty value.
- Segment variables not starting immediately after a path separator may have
empty values.

- When matching an end variable, the value matches until the end of the request
path.
    - If the end variable value would be the empty string, and if the static route
    before the end variable is servable, then the static route is served. Otherwise,
    the end variable route is used, which may or may not be servable,
    which may or may not result in an error handler.

- If a trailing path separator is allowed, then when a route is found,
and the remaining path to find is empty or exactly `/`, then that route is served. Otherwise,
the ErrNotFound handler is used.
- If a trailing path separator is not allowed, then when a route is found, and
the remaining path is not the empty string, then the ErrNotFound handler is used.
