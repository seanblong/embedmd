# CR031 - Fenced code blocks should be surrounded by blank lines

Tags: code, blank_lines

Aliases: blanks-around-fences

This rule is triggered when fenced code blocks are either not preceded or not
followed by a blank line:

    Some text
    ```
    Code block
    ```

    ```
    Another code block
    ```
    Some more text

To fix this, ensure that all fenced code blocks have a blank line both before
and after (except where the block is at the beginning or end of the document):

    Some text

    ```
    Code block
    ```

    ```
    Another code block
    ```

    Some more text

Rationale: Aside from aesthetic reasons, some parsers, including kramdown, will
not parse fenced code blocks that don't have blank lines before and after them.

An exception can be made for certain prefixes such as comments or other metadata.
To ignore these lines in the rule, leverage the `ignore_prefix` parameter that allows
you to ignore lines beginning with a specific prefix, e.g., "<!-" or "[embedmd]"
