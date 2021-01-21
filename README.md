# desktop-audio

An experiment to play sounds with Web UI and _without_ Web Audio API.

## Architecture

```
┌──────────────── Electron ────────────────┐
┌──────────────┐           ┌───────────────┐           ┌───────────┐
│ ui(TS/React) │<-- IPC -->│ core(TS/Node) │<-- IPC -->│ audio(Go) │<--- MIDI
└──────────────┘           └───────────────┘           └───────────┘
```

## IPC Protocol

```
{url-encoded} {url-encoded} {url-encoded} ...
{url-encoded} {url-encoded} {url-encoded} ...
...
```

e.g. `note_on 60`
