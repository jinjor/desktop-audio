# desktop-audio

An experiment to play sounds with Web UI and _without_ Web Audio API.

## Architecture

```
┌───────────── Electron ─────────────┐
┌────────────┐           ┌───────────┐           ┌───────┐
│ ui         │<-- IPC -->│ core      │<-- IPC -->│ audio │<---- MIDI
│ (TS/React) │           │ (TS/Node) │           │ (Go)  │<---> File
└────────────┘           └───────────┘           └───────┘

Loops:
                 MIDI
┌──────┐      .   │    ┌───>──┐     ┌───>───┐
│      │      .   └───>│ midi │---->│       │
│      │      .        └──<───┘     │       │
│      │      .        ┌───>──┐     │       │
│      │---- IPC ----->│ recv │---->│       │
│ core │      .        └──<───┘     │ audio │
│      │      .        ┌───>──┐     │       │
│      │<--- IPC ------│ send │<----│       │
│      │      .        └──<───┘     │       │
│      │      .        ┌───>──┐     │       │
│      │      .   ┌─<──│ save │<----│       │
└──────┘      .   │    └──<───┘     └───<───┘
                 File
```

## IPC Protocol

```
{url-encoded} {url-encoded} {url-encoded} ...
{url-encoded} {url-encoded} {url-encoded} ...
...
```

e.g. `note_on 60`
