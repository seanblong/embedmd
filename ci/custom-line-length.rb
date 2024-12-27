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
