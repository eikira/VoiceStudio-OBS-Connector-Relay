' Stops the relay and removes its notification-area icon.
Set sh = CreateObject("WScript.Shell")
Set fso = CreateObject("Scripting.FileSystemObject")
dir = fso.GetParentFolderName(WScript.ScriptFullName)
exe = "VoiceStudio-OBS-Connector-Relay.exe"
If sh.ExpandEnvironmentStrings("%PROCESSOR_ARCHITECTURE%") = "x86" And sh.ExpandEnvironmentStrings("%PROCESSOR_ARCHITEW6432%") = "%PROCESSOR_ARCHITEW6432%" Then exe = "VoiceStudio-OBS-Connector-Relay-32bit.exe"
code = sh.Run("""" & dir & "\" & exe & """ --stop", 0, True)
If code = 0 Then
  sh.Popup "The relay has been stopped.", 2, "VoiceStudio OBS Connector Relay", 64
Else
  sh.Popup "The relay was not running.", 2, "VoiceStudio OBS Connector Relay", 64
End If
