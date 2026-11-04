Run

```sh
go run .
```

Install


```sh
go build
cp notepad.exe "$LOCALAPPDATA/Microsoft/WindowsApps/notepad.exe"
```

## Owning the `notepad` name

Windows builds a process's `PATH` as the system `PATH` first, then the user
`PATH` appended. `C:\Windows\System32` is a system entry, while
`%LOCALAPPDATA%\Microsoft\WindowsApps` is a user entry, so a binary installed
there is searched *after* the built-in editor and never wins. Inside Git Bash,
MSYS also prepends its own `/usr/bin`, whose `notepad` shim wins first.

To make this tool the `notepad` that runs, override it per shell. Shell
functions and aliases take precedence over `PATH`.

Bash (`~/.bashrc`):

```sh
notepad() { "$LOCALAPPDATA/Microsoft/WindowsApps/notepad.exe" "$@"; }
```

PowerShell (`$PROFILE`):

```powershell
function notepad { & "$env:LOCALAPPDATA\Microsoft\WindowsApps\notepad.exe" @args }
```

cmd (AutoRun doskey macro):

```cmd
doskey notepad="%LOCALAPPDATA%\Microsoft\WindowsApps\notepad.exe" $*
```

For a system-wide override (Run dialog and every process), move the install
directory to the front of the system `Path`, ahead of `%SystemRoot%\System32`
(requires admin). Git Bash still prefers its own `/usr/bin/notepad`, so keep
the bash function above as well.

The real editor is left untouched on disk (needed for Windows servicing); only
the `notepad` command is taken over.


