# cmdperf demo tape

Output doc/demo.gif
Set FontFamily "IBM Plex Mono"
#Set FontFamily "0xProto"

Set Padding 10
# Set up shell appearance
Set Shell "zsh"
Set FontSize 16
Set Width 960
Set Height 500
Set PlaybackSpeed 0.2
Set Framerate 60

Type 'cmdperf -r 1 "curl -s google.com" "curl -s bing.com"' Sleep 200ms
Enter
Sleep 12s

Type 'cmdperf -r 1 "curl -s google.com" "curl -s bing.com"' Sleep 200ms
Enter
Sleep 12s

# End of demo
Sleep 2s
