# drawdeck

Draw cards at random from markdown checklists. A terminal app built with [Bubble Tea](https://charm.land/bubbletea).

Point it at a markdown file full of checkboxes and it becomes a deck. Drawing picks an unchecked card at random, checks it off in the file, and remembers it — so the file stays a plain, hand-editable document you own.

## Install

```sh
go build -o drawdeck
```

## Usage

```sh
drawdeck                    # open the deck list
drawdeck add DECK.md        # register a deck and open it
drawdeck -h                 # help
drawdeck -v                 # version
```

## Keys

**List view**

| Key | Action |
| --- | ------ |
| `↑↓` / `jk` | navigate |
| `enter` / `l` | open deck |
| `a` | add a deck by path |
| `d` | remove from list (file untouched) |
| `r` | re-read every deck file |
| `q` | quit |

**Deck view**

| Key | Action |
| --- | ------ |
| hold `space` | draw a card |
| `space` ×3 | draw a card, any terminal |
| `ctrl+r` | reset the deck |
| `esc` | back to list |

Hold-to-draw needs a terminal supporting the Kitty keyboard protocol (Ghostty, Kitty, WezTerm). Everywhere else, tapping `space` three times does the same job — the prompt tells you which is active.

## Deck files

The top `#` header names the deck. Every non-indented checkbox is a card; an indented quote beneath it becomes that card's notes.

```markdown
# Example Deck

- [ ] A woolly yak
- [ ] A giant elephant
    > This one is my favorite
```

Drawing checks the card off and records it in frontmatter at the top of the file. Every other byte is left exactly as you wrote it — prose, extra headers, and nested lists all survive.

Decks are registered in `$XDG_CONFIG_HOME/drawdeck/decks.json` (or `~/.config/drawdeck/decks.json`), which stores paths only.
