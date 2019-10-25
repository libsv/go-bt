const bsv = require('bsv')

const utxo = bsv.Transaction.UnspentOutput({
  txid: '63daf8b0961867e8f2fad1c04a4ceed2618a33aa73bae1c9c8540e9484ed7d03',
  vout: 0,
  scriptPubKey: '76a91403ececf2d12a7f614aef4c82ecf13c303bd9975d88ac',
  amount: 49.98000000
})

const tx = bsv.Transaction('0200000001037ded84940e54c8c9e1ba73aa338a61d2ee4c4ac0d1faf2e8671896b0f8da630000000000ffffffff01806de729010000001976a91463ea0d776d45502d2226aed9ebdf5b676e232ca188ac00000000')
tx.inputs = []
tx.from(utxo)
const privKey = bsv.PrivateKey.fromWIF('cPjqbeH84Qq9VmWrURUEJNo7DaKnrPP428utXzZRcbBdXPx7kGe5')

tx.sign(privKey)

console.log(tx.toObject())
console.log(tx.isFullySigned())

// ee27ec55cbdcfa87d01ba83244593aef2a50dabc2dc4eb7f2bd2a5d04a76d0f9
