# zed-ext-install

I wanted to try the [Zed](https://zed.dev) editor for a long time, but kept putting it off because extensions just wouldn't install. Every time I tried — same error:

```
ERROR downloading extension
Caused by:
    connection error: peer closed connection without sending TLS close_notify
```

I'm not 100% sure why this happens, but it seems like Zed downloads extensions from DigitalOcean Spaces, and the connection just drops sometimes. And as far as I can tell, there is no retry logic, so you have to do it again and again by hand.

So I made this small CLI tool in Go that does the same thing but with automatic retries.

If you have the same problem — feel free to use it.

---

## The problem

![Extension install failing in Zed GUI](assets/toml_error_install.gif)

## The fix

```sh
zed-ext-install install toml
```

![Installing via zed-ext-install in terminal](assets/toml_zei_install.gif)

After that, restart Zed and the extension will appear as installed:

![Extension showing as installed in Zed](assets/toml_success_install.png)

---

## Install

```sh
git clone https://github.com/skyvxl/zed-ext-install
cd zed-ext-install
go build -o zed-ext-install .
```

Or just move the binary somewhere in your `$PATH`.

## Usage

```sh
# Search for an extension
zed-ext-install search html

# Install an extension (latest version)
zed-ext-install install html

# Install a specific version
zed-ext-install install html 0.3.0

# List installed extensions
zed-ext-install list

# Remove an extension
zed-ext-install remove html
```

## How it works

Zed downloads extension archives from DigitalOcean Spaces via a pre-signed URL that expires in 3 minutes. If the connection drops, Zed just fails with no retry.

This tool does the same download but retries up to 5 times with exponential backoff. It also unpacks the archive and updates `index.json` so Zed picks up the extension on next start.

## Supported platforms

- macOS (`~/Library/Application Support/Zed/extensions/`)
- Linux (`~/.local/share/zed/extensions/`) (no idea, didn't test, but should work probably))

## Contributing

If you have any ideas or suggestions — feel free to open an issue or a pull request. Always welcome.
