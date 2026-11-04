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

Copying into `WindowsApps` makes this binary shadow the built-in editor,
because that folder comes before `C:\Windows\System32` in `PATH`.

On Windows 11 the Store Notepad also registers an App Execution Alias that can
intercept the name. Disable it so this tool is the only `notepad`:

Settings → Apps → Advanced app settings → App execution aliases → turn
**Notepad** off.

The real editor is left untouched on disk (needed for Windows servicing); only
the `notepad` command is taken over.

