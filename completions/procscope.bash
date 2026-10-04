# bash completion for procscope

_procscope() {
    local cur prev opts
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"

    opts="--pid -p --name -n --out -o --jsonl --summary --no-color --quiet -q --max-args --max-path --skip-checks --version --help completion"

    case "${prev}" in
        -p|--pid)
            # Complete PIDs from /proc
            COMPREPLY=( $(compgen -W "$(ls /proc 2>/dev/null | grep -E '^[0-9]+$')" -- "${cur}") )
            return 0
            ;;
        -o|--out|--jsonl|--summary)
            # Complete file/directory paths
            COMPREPLY=( $(compgen -f -- "${cur}") )
            return 0
            ;;
        -n|--name)
            # Complete process names from /proc/*/comm
            COMPREPLY=( $(compgen -W "$(cat /proc/*/comm 2>/dev/null | sort -u)" -- "${cur}") )
            return 0
            ;;
        --max-args|--max-path)
            return 0
            ;;
    esac

    if [[ "${cur}" == -* ]]; then
        COMPREPLY=( $(compgen -W "${opts}" -- "${cur}") )
        return 0
    fi
}

complete -F _procscope procscope
