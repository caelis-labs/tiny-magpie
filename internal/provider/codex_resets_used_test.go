package provider

import ()

// foobra12 on Discord: "magpie会自动使用codex重置卡吗 我发现刚才好像不小心
// 自动给我使用了一张". A reset magpie spent by itself was said only in the
// gateway's log, so a reset gone from the card read the same whether magpie
// used it, the user did, or it ran out. Each one magpie spends is kept, with
// why, and the account's reading carries them, newest first: one spent as
// it was about to run out, one by hand, and one as the week was used up.
