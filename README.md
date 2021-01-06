# desktop-audio

An experiment to play sounds with Web UI and without Web Audio API.

## Architecture

```
┌──────────────── Electron ────────────────┐
┌──────────────┐           ┌───────────────┐           ┌───────────┐
│ ui(TS/React) │<-- IPC -->│ core(TS/Node) │<-- IPC -->│ audio(Go) │
└──────────────┘           │               │           └───────────┘
                           │               │       ┌─────────────┐
                           │               │<----->│ File System │
                           └───────────────┘       └─────────────┘
```

## IPC Protocol

```
{url-encoded} {url-encoded} {url-encoded} ...
{url-encoded} {url-encoded} {url-encoded} ...
...
```

e.g. `note_on 60`
