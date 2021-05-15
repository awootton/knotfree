package tokens

// SampleSmallToken is a small token signed by "_9sh" (below)
// p.Input = 20
// p.Output = 20
// p.Subscriptions = 2
// p.Connections = 2
var SampleSmallToken = `[My_token_expires:_2021-12-31,{exp:1641023999,iss:_9sh,jti:amXYKIuS4uykvPem9Fml371o,in:32,out:32,su:4,co:2,url:knotfree.net},eyJhbGciOiJFZDI1NTE5IiwidHlwIjoiSldUIn0.eyJleHAiOjE2NDEwMjM5OTksImlzcyI6Il85c2giLCJqdGkiOiJhbVhZS0l1UzR1eWt2UGVtOUZtbDM3MW8iLCJpbiI6MzIsIm91dCI6MzIsInN1Ijo0LCJjbyI6MiwidXJsIjoia25vdGZyZWUubmV0In0.7ElPyX1Vju8Q5uDOEfgAblwvE2gxT78Jl68JPlqLRcFeMJ7if39Ppl2_Jr_JTky371bIXAn6S-pghtWSqTBwAQ]`

// no point loading them all the time.
// ed25519
// one per line.
// _9sh is being used to sign tokens
// 8ZNP is unused
// yRst is used as seed to cluster box keypair
// the others are unused so far and the private part unloaded.
var PublicKeys string = `
_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
`
