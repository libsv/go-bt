package bt

import "github.com/pkg/errors"

// General errors.
var (
	ErrInvalidTxID       = errors.New("invalid TxID")
	ErrTxNil             = errors.New("tx is nil")
	ErrTxTooShort        = errors.New("too short to be a tx - even an empty tx has 10 bytes")
	ErrNLockTimeLength   = errors.New("nLockTime length must be 4 bytes long")
	ErrEmptyValues       = errors.New("empty value or values passed, all arguments are required and cannot be empty")
	ErrUnsupportedScript = errors.New("non-P2PKH input used in the tx - unsupported")
	ErrInvalidScriptType = errors.New("invalid script type")
	ErrNoUnlocker        = errors.New("unlocker not supplied")
)

// Sentinal errors reported by inputs.
var (
	ErrInputNoExist  = errors.New("specified input does not exist")
	ErrInputTooShort = errors.New("input length too short")

	// You should not be able to spend an input with 0 Satoshi value.
	// Most likely the input Satoshi value is not provided.
	ErrInputSatsZero = errors.New("input satoshi value is not provided")
)

// Sentinal errors reported by outputs.
var (
	ErrOutputNoExist  = errors.New("specified output does not exist")
	ErrOutputTooShort = errors.New("output length too short")
)

// Sentinal errors reported by change.
var (
	ErrInsufficientInputs = errors.New("satoshis inputted to the tx are less than the outputted satoshis")
)

// Sentinal errors reported by signature hash.
var (
	ErrEmptyPreviousTxID     = errors.New("'PreviousTxID' not supplied")
	ErrEmptyPreviousTxScript = errors.New("'PreviousTxScript' not supplied")
)

// Sentinel errors reported by the fees.
var (
	ErrFeeQuotesNotInit = errors.New("feeQuotes have not been setup, call NewFeeQuotes")
	ErrMinerNoQuotes    = errors.New("miner has no quotes stored")
	ErrFeeTypeNotFound  = errors.New("feetype not found")
	ErrFeeQuoteNotInit  = errors.New("feeQuote has not been initialised, call NewFeeQuote()")
	ErrUnknownFeeType   = errors.New("unknown fee type")
)

// Sentinel errors reported by Fund.
var (
	// ErrNoUTXO signals the UTXOGetterFunc has reached the end of its input.
	ErrNoUTXO = errors.New("no remaining utxos")

	// ErrInsufficientFunds insufficient funds provided for funding
	ErrInsufficientFunds = errors.New("insufficient funds provided")
)

// Sentinal errors reported by ordinal inscriptions.
var (
	ErrOutputsNotEmpty = errors.New("transaction outputs must be empty to avoid messing with Ordinal ordering scheme")
)

// Sentinal errors reported by PSBTs.
var (
	ErrDummyInput           = errors.New("failed to add dummy input 0")
	ErrInsufficientUTXOs    = errors.New("need at least 3 utxos")
	ErrUTXOInputMismatch    = errors.New("utxo and input mismatch")
	ErrInvalidSellOffer     = errors.New("invalid sell offer (partially signed tx)")
	ErrOrdinalOutputNoExist = errors.New("ordinal output expected in index 2 doesn't exist")
	ErrOrdinalInputNoExist  = errors.New("ordinal input expected in index 2 doesn't exist")
)
