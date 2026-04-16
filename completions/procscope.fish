# Fish completion for procscope

complete -c procscope -s p -l pid -d "Attach to existing process by PID" -x -a "(__fish_complete_pids)"
complete -c procscope -s n -l name -d "Attach to process by name"
complete -c procscope -s o -l out -d "Evidence bundle output directory" -r -F
complete -c procscope -l jsonl -d "Write events as JSONL to file" -r -F
complete -c procscope -l summary -d "Write Markdown summary to file" -r -F
complete -c procscope -l no-color -d "Disable colored output"
complete -c procscope -s q -l quiet -d "Suppress live timeline"
complete -c procscope -l max-args -d "Maximum argv elements" -x
complete -c procscope -l max-path -d "Maximum path length" -x
complete -c procscope -l skip-checks -d "Skip privilege checks"
complete -c procscope -l version -d "Show version"
complete -c procscope -s h -l help -d "Show help"

# Subcommands
complete -c procscope -n "__fish_use_subcommand" -a completion -d "Generate shell completions"
complete -c procscope -n "__fish_seen_subcommand_from completion" -a "bash zsh fish"
