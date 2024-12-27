# CR013 - Custom line length

Tags: line_length

Aliases: custom-line-length

Parameters: line_length, ignore_code_blocks, code_blocks, tables, ignore_prefix
(number; default 80, boolean; default false, boolean; default true, boolean; default
true, string; default nil)

This rule is triggered when there are lines that are longer than the
configured line length (default: 80 characters). To fix this, split the line
up into multiple lines.

This rule has an exception where there is no whitespace beyond the configured
line length. This allows you to still include items such as long URLs without
being forced to break them in the middle.

You also have the option to exclude this rule for code blocks. To
do this, set the `ignore_code_blocks` parameter to true. To exclude this rule
for tables set the `tables` parameters to false.  Setting the parameter
`code_blocks` to false to exclude the rule for code blocks is deprecated and
will be removed in a future release.

This custom rule also includes an `ignore_prefix` parameter that allows you to
ignore lines beginning with a specific prefix. This is useful for ignoring long
lines in Markdown files that are used for embedding code snippets. For example,
to ignore lines that begin with `[embedmd]` you would set the `ignore_prefix`
parameter to `[embedmd]`.

Code blocks are included in this rule by default since it is often a
requirement for document readability, and tentatively compatible with code
rules. Still, some languages do not lend themselves to short lines.
