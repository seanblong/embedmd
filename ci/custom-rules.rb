rule 'CR013', 'Custom line length' do
    tags :line_length
    aliases 'custom-line-length'
    docs 'https://github.com/seanblong/embedmd/blob/main/ci/custom-line-length.md'
    params :line_length => 80, :ignore_code_blocks => false, :code_blocks => true,
           :tables => true, :ignore_prefix => nil

    check do |doc|
      # Every line in the document that is part of a code block.
      codeblock_lines = doc.find_type_elements(:codeblock).map do |e|
        (doc.element_linenumber(e)..
                 doc.element_linenumber(e) + e.value.lines.count).to_a
      end.flatten
      # Every line in the document that is part of a table.
      locations = doc.elements
                     .map { |e| [e.options[:location], e] }
                     .reject { |l, _| l.nil? }
      table_lines = locations.map.with_index do |(l, e), i|
        if e.type == :table
          if i + 1 < locations.size
            (l..locations[i + 1].first - 1).to_a
          else
            (l..doc.lines.count).to_a
          end
        end
      end.flatten
      overlines = doc.matching_lines(/^.{#{@params[:line_length]}}.*\s/)
      if !params[:code_blocks] || params[:ignore_code_blocks]
        overlines -= codeblock_lines
        unless params[:code_blocks]
          warn 'MD013 warning: Parameter :code_blocks is deprecated.'
          warn '  Please replace \":code_blocks => false\" by '\
               '\":ignore_code_blocks => true\" in your configuration.'
        end
      end
      if !params[:ignore_prefix].nil?
        overlines -= doc.matching_lines(/^.#{params[:ignore_prefix]}.*\s/)
      end
      overlines -= table_lines unless params[:tables]
      overlines
    end
  end

rule 'CR031', 'Custom fenced code blocks should be surrounded by blank lines' do
    tags :code, :blank_lines
    aliases 'custom-blanks-around-fences'
    docs 'https://github.com/seanblong/embedmd/blob/main/ci/custom-fenced-code-blocks.md'
    params :ignore_prefix => nil
    check do |doc|
      errors = []
      # Some parsers (including kramdown) have trouble detecting fenced code
      # blocks without surrounding whitespace, so examine the lines directly.
      in_code = false
      fence = nil
      lines = [''] + doc.lines + ['']
      ignore_prefix = params[:ignore_prefix]

      # Define a helper to check if a line starts with the ignore_prefix
      should_ignore = lambda do |line, prefix|
        return false if prefix.nil? || prefix.empty?
        # If prefix is an array, check any of the prefixes
        prefixes = prefix.is_a?(Array) ? prefix : [prefix]
        prefixes.any? { |p| line.strip.start_with?(p) }
      end

      lines.each_with_index do |line, linenum|
        line.strip.match(/^(`{3,}|~{3,})/)
        unless Regexp.last_match(1) &&
               (
                 !in_code ||
                 (Regexp.last_match(1).slice(0, fence.length) == fence)
               )
          next
        end

        fence = in_code ? nil : Regexp.last_match(1)
        in_code = !in_code

        # Retrieve adjacent lines
        previous_line = lines[linenum - 1]
        next_line = lines[linenum + 1]

        # Determine if the adjacent lines violate the blank line requirement
        previous_issue = in_code && !previous_line.strip.empty? && !should_ignore.call(previous_line, ignore_prefix)
        next_issue = !in_code && !next_line.strip.empty? && !should_ignore.call(next_line, ignore_prefix)

        # Record error if any issue is found
        if previous_issue || next_issue
          errors << linenum
        end
      end
      errors
    end
  end