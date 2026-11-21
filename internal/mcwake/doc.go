// Package mcwake implements the small, offline part of the Minecraft Java TCP
// protocol needed to advertise a stopped server and request a wake on login.
// It never authenticates a player, establishes a game session, dials a server,
// or forwards traffic. The handshake host is parsed only, never used for routing.
//
// Status requests echo the client's nonnegative protocol number. Login supports
// release protocols 4 through 758 (1.7.2 through 1.18.2, username only), 759
// (1.19, optional signature), 760 (1.19.1/2, optional signature and UUID),
// 761 through 763 (1.19.3 through 1.20.1, optional UUID), and 764 and newer
// (1.20.2 onward, required UUID). Newer versions are accepted only while they
// retain that Login Start shape. Transfer state 3 is treated as login only for
// protocol 766 and newer. Pre-1.7 and legacy ping formats are not supported.
// Signature fields are bounded and structurally parsed, not authenticated;
// neither the username nor UUID is an authenticated identity.
//
// Login disconnects always contain a length-prefixed JSON text component string,
// including on 1.21.6; they are not NBT. Version-specific Login Start layouts
// follow PrismarineJS minecraft-data's data/pc/{1.19,1.19.2,1.19.3,1.19.4,
// 1.20,1.20.2,1.21.6}/protocol.json at
// https://github.com/PrismarineJS/minecraft-data/tree/master/data/pc.
//
// Input frames are limited to 8192 bytes. VarInts must be canonical,
// nonnegative signed-32-bit values. Hosts are valid UTF-8 with at most 255
// UTF-16 code units; usernames contain 1-16 ASCII letters, digits or underscores.
// The complete status response is limited to 128 KiB on the wire. Disconnect
// JSON strings are limited to 4096 bytes; an oversized callback message is
// replaced by a short, friendly fallback. Parent-supplied status fields are
// JSON escaped, with text, newlines and color codes otherwise preserved. Icon
// data URLs must be validated by the caller: this package does not fetch URLs
// or decode images.
//
// Handle owns its connection and sets one absolute five-second read/write
// deadline, without extending it per packet. Callbacks run synchronously and
// must return promptly; the I/O deadline cannot interrupt a blocking callback.
// Invalid input never triggers a wake. Each valid login calls join exactly once,
// even if the subsequent disconnect write fails. Concurrent Handle calls are
// independent; global wake deduplication belongs to the caller.
package mcwake
