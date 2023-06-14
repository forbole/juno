package types_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/forbole/juno/v5/types"
)

func TestTransactionUnmarshalling(t *testing.T) {
	transactionData := `
{
  "height": "8968076",
  "txhash": "A24B8CC71DB53E0AF50734F38F0A2BE40F21C315D21C4087FABB8833B8FD08D0",
  "codespace": "",
  "code": 0,
  "data": "0A230A1D2F636F736D6F732E617574687A2E763162657461312E4D73674578656312020A00",
  "raw_log": "[{\"events\":[{\"type\":\"coin_received\",\"attributes\":[{\"key\":\"receiver\",\"value\":\"desmos19jhmt3f49fdr4tev2lnux20g0qjetd4lnd6xmy\"},{\"key\":\"amount\",\"value\":\"101031udsm\"},{\"key\":\"authz_msg_index\",\"value\":\"0\"},{\"key\":\"receiver\",\"value\":\"desmos1fl48vsnmsdzcv85q5d2q4z5ajdha8yu3prylw0\"},{\"key\":\"amount\",\"value\":\"100977udsm\"},{\"key\":\"authz_msg_index\",\"value\":\"0\"}]},{\"type\":\"coin_spent\",\"attributes\":[{\"key\":\"spender\",\"value\":\"desmos1jv65s3grqf6v6jl3dp4t6c9t9rk99cd8n8fv78\"},{\"key\":\"amount\",\"value\":\"101031udsm\"},{\"key\":\"authz_msg_index\",\"value\":\"0\"},{\"key\":\"spender\",\"value\":\"desmos19jhmt3f49fdr4tev2lnux20g0qjetd4lnd6xmy\"},{\"key\":\"amount\",\"value\":\"100977udsm\"},{\"key\":\"authz_msg_index\",\"value\":\"0\"}]},{\"type\":\"delegate\",\"attributes\":[{\"key\":\"validator\",\"value\":\"desmosvaloper1zngdx77g9ywnwmwpwvj9w2eqcs6fhw78gn02d8\"},{\"key\":\"amount\",\"value\":\"100977udsm\"},{\"key\":\"new_shares\",\"value\":\"100987.098700553264185967\"},{\"key\":\"authz_msg_index\",\"value\":\"0\"}]},{\"type\":\"message\",\"attributes\":[{\"key\":\"action\",\"value\":\"/cosmos.authz.v1beta1.MsgExec\"},{\"key\":\"sender\",\"value\":\"desmos1jv65s3grqf6v6jl3dp4t6c9t9rk99cd8n8fv78\"},{\"key\":\"authz_msg_index\",\"value\":\"0\"},{\"key\":\"module\",\"value\":\"staking\"},{\"key\":\"sender\",\"value\":\"desmos19jhmt3f49fdr4tev2lnux20g0qjetd4lnd6xmy\"},{\"key\":\"authz_msg_index\",\"value\":\"0\"}]},{\"type\":\"transfer\",\"attributes\":[{\"key\":\"recipient\",\"value\":\"desmos19jhmt3f49fdr4tev2lnux20g0qjetd4lnd6xmy\"},{\"key\":\"sender\",\"value\":\"desmos1jv65s3grqf6v6jl3dp4t6c9t9rk99cd8n8fv78\"},{\"key\":\"amount\",\"value\":\"101031udsm\"},{\"key\":\"authz_msg_index\",\"value\":\"0\"}]},{\"type\":\"withdraw_rewards\",\"attributes\":[{\"key\":\"amount\",\"value\":\"101031udsm\"},{\"key\":\"validator\",\"value\":\"desmosvaloper1zngdx77g9ywnwmwpwvj9w2eqcs6fhw78gn02d8\"},{\"key\":\"authz_msg_index\",\"value\":\"0\"}]}]}]",
  "logs": [
    {
      "msg_index": 0,
      "log": "",
      "events": [
        {
          "type": "coin_received",
          "attributes": [
            {
              "key": "receiver",
              "value": "desmos19jhmt3f49fdr4tev2lnux20g0qjetd4lnd6xmy"
            },
            {
              "key": "amount",
              "value": "101031udsm"
            },
            {
              "key": "authz_msg_index",
              "value": "0"
            },
            {
              "key": "receiver",
              "value": "desmos1fl48vsnmsdzcv85q5d2q4z5ajdha8yu3prylw0"
            },
            {
              "key": "amount",
              "value": "100977udsm"
            },
            {
              "key": "authz_msg_index",
              "value": "0"
            }
          ]
        },
        {
          "type": "coin_spent",
          "attributes": [
            {
              "key": "spender",
              "value": "desmos1jv65s3grqf6v6jl3dp4t6c9t9rk99cd8n8fv78"
            },
            {
              "key": "amount",
              "value": "101031udsm"
            },
            {
              "key": "authz_msg_index",
              "value": "0"
            },
            {
              "key": "spender",
              "value": "desmos19jhmt3f49fdr4tev2lnux20g0qjetd4lnd6xmy"
            },
            {
              "key": "amount",
              "value": "100977udsm"
            },
            {
              "key": "authz_msg_index",
              "value": "0"
            }
          ]
        },
        {
          "type": "delegate",
          "attributes": [
            {
              "key": "validator",
              "value": "desmosvaloper1zngdx77g9ywnwmwpwvj9w2eqcs6fhw78gn02d8"
            },
            {
              "key": "amount",
              "value": "100977udsm"
            },
            {
              "key": "new_shares",
              "value": "100987.098700553264185967"
            },
            {
              "key": "authz_msg_index",
              "value": "0"
            }
          ]
        },
        {
          "type": "message",
          "attributes": [
            {
              "key": "action",
              "value": "/cosmos.authz.v1beta1.MsgExec"
            },
            {
              "key": "sender",
              "value": "desmos1jv65s3grqf6v6jl3dp4t6c9t9rk99cd8n8fv78"
            },
            {
              "key": "authz_msg_index",
              "value": "0"
            },
            {
              "key": "module",
              "value": "staking"
            },
            {
              "key": "sender",
              "value": "desmos19jhmt3f49fdr4tev2lnux20g0qjetd4lnd6xmy"
            },
            {
              "key": "authz_msg_index",
              "value": "0"
            }
          ]
        },
        {
          "type": "transfer",
          "attributes": [
            {
              "key": "recipient",
              "value": "desmos19jhmt3f49fdr4tev2lnux20g0qjetd4lnd6xmy"
            },
            {
              "key": "sender",
              "value": "desmos1jv65s3grqf6v6jl3dp4t6c9t9rk99cd8n8fv78"
            },
            {
              "key": "amount",
              "value": "101031udsm"
            },
            {
              "key": "authz_msg_index",
              "value": "0"
            }
          ]
        },
        {
          "type": "withdraw_rewards",
          "attributes": [
            {
              "key": "amount",
              "value": "101031udsm"
            },
            {
              "key": "validator",
              "value": "desmosvaloper1zngdx77g9ywnwmwpwvj9w2eqcs6fhw78gn02d8"
            },
            {
              "key": "authz_msg_index",
              "value": "0"
            }
          ]
        }
      ]
    }
  ],
  "info": "",
  "gas_wanted": "213372",
  "gas_used": "195395",
  "tx": {
    "@type": "/cosmos.tx.v1beta1.Tx",
    "body": {
      "messages": [
        {
          "@type": "/cosmos.authz.v1beta1.MsgExec",
          "grantee": "desmos1pu0pmlvakgdn7lzrwx2rgj6xjvcyln0l4cdfl2",
          "msgs": [
            {
              "@type": "/cosmos.staking.v1beta1.MsgDelegate",
              "delegator_address": "desmos19jhmt3f49fdr4tev2lnux20g0qjetd4lnd6xmy",
              "validator_address": "desmosvaloper1zngdx77g9ywnwmwpwvj9w2eqcs6fhw78gn02d8",
              "amount": {
                "denom": "udsm",
                "amount": "100977"
              }
            }
          ]
        }
      ],
      "memo": "REStaked by Stakewolle.com |  Auto-compound",
      "timeout_height": "0",
      "extension_options": [],
      "non_critical_extension_options": []
    },
    "auth_info": {
      "signer_infos": [
        {
          "public_key": {
            "@type": "/cosmos.crypto.secp256k1.PubKey",
            "key": "Aru+kPlD8wXCTcBO/Gw4OniqGKFCZdrva7fjtbY3phUf"
          },
          "mode_info": {
            "single": {
              "mode": "SIGN_MODE_DIRECT"
            }
          },
          "sequence": "6655"
        }
      ],
      "fee": {
        "amount": [
          {
            "denom": "udsm",
            "amount": "5335"
          }
        ],
        "gas_limit": "213372",
        "payer": "",
        "granter": ""
      },
      "tip": null
    },
    "signatures": [
      "BYjVmo5qvRqcxtkMH2k/jPxbsZI67TKSprEH1/oYmypT1exfJ3gFAEhaQQwOVf2AnDaYXnn+QHymhYcHz7r6Ww=="
    ]
  },
  "timestamp": "2023-06-14T19:12:57Z",
  "events": [
    {
      "type": "coin_spent",
      "attributes": [
        {
          "key": "c3BlbmRlcg==",
          "value": "ZGVzbW9zMXB1MHBtbHZha2dkbjdsenJ3eDJyZ2o2eGp2Y3lsbjBsNGNkZmwy",
          "index": true
        },
        {
          "key": "YW1vdW50",
          "value": "NTMzNXVkc20=",
          "index": true
        }
      ]
    },
    {
      "type": "coin_received",
      "attributes": [
        {
          "key": "cmVjZWl2ZXI=",
          "value": "ZGVzbW9zMTd4cGZ2YWttMmFtZzk2MnlsczZmODR6M2tlbGw4YzVseXB3c3U5",
          "index": true
        },
        {
          "key": "YW1vdW50",
          "value": "NTMzNXVkc20=",
          "index": true
        }
      ]
    },
    {
      "type": "transfer",
      "attributes": [
        {
          "key": "cmVjaXBpZW50",
          "value": "ZGVzbW9zMTd4cGZ2YWttMmFtZzk2MnlsczZmODR6M2tlbGw4YzVseXB3c3U5",
          "index": true
        },
        {
          "key": "c2VuZGVy",
          "value": "ZGVzbW9zMXB1MHBtbHZha2dkbjdsenJ3eDJyZ2o2eGp2Y3lsbjBsNGNkZmwy",
          "index": true
        },
        {
          "key": "YW1vdW50",
          "value": "NTMzNXVkc20=",
          "index": true
        }
      ]
    },
    {
      "type": "message",
      "attributes": [
        {
          "key": "c2VuZGVy",
          "value": "ZGVzbW9zMXB1MHBtbHZha2dkbjdsenJ3eDJyZ2o2eGp2Y3lsbjBsNGNkZmwy",
          "index": true
        }
      ]
    },
    {
      "type": "tx",
      "attributes": [
        {
          "key": "ZmVl",
          "value": "NTMzNXVkc20=",
          "index": true
        },
        {
          "key": "ZmVlX3BheWVy",
          "value": "ZGVzbW9zMXB1MHBtbHZha2dkbjdsenJ3eDJyZ2o2eGp2Y3lsbjBsNGNkZmwy",
          "index": true
        }
      ]
    },
    {
      "type": "tx",
      "attributes": [
        {
          "key": "YWNjX3NlcQ==",
          "value": "ZGVzbW9zMXB1MHBtbHZha2dkbjdsenJ3eDJyZ2o2eGp2Y3lsbjBsNGNkZmwyLzY2NTU=",
          "index": true
        }
      ]
    },
    {
      "type": "tx",
      "attributes": [
        {
          "key": "c2lnbmF0dXJl",
          "value": "QllqVm1vNXF2UnFjeHRrTUgyay9qUHhic1pJNjdUS1NwckVIMS9vWW15cFQxZXhmSjNnRkFFaGFRUXdPVmYyQW5EYVlYbm4rUUh5bWhZY0h6N3I2V3c9PQ==",
          "index": true
        }
      ]
    },
    {
      "type": "message",
      "attributes": [
        {
          "key": "YWN0aW9u",
          "value": "L2Nvc21vcy5hdXRoei52MWJldGExLk1zZ0V4ZWM=",
          "index": true
        }
      ]
    },
    {
      "type": "coin_spent",
      "attributes": [
        {
          "key": "c3BlbmRlcg==",
          "value": "ZGVzbW9zMWp2NjVzM2dycWY2djZqbDNkcDR0NmM5dDlyazk5Y2Q4bjhmdjc4",
          "index": true
        },
        {
          "key": "YW1vdW50",
          "value": "MTAxMDMxdWRzbQ==",
          "index": true
        },
        {
          "key": "YXV0aHpfbXNnX2luZGV4",
          "value": "MA==",
          "index": true
        }
      ]
    },
    {
      "type": "coin_received",
      "attributes": [
        {
          "key": "cmVjZWl2ZXI=",
          "value": "ZGVzbW9zMTlqaG10M2Y0OWZkcjR0ZXYybG51eDIwZzBxamV0ZDRsbmQ2eG15",
          "index": true
        },
        {
          "key": "YW1vdW50",
          "value": "MTAxMDMxdWRzbQ==",
          "index": true
        },
        {
          "key": "YXV0aHpfbXNnX2luZGV4",
          "value": "MA==",
          "index": true
        }
      ]
    },
    {
      "type": "transfer",
      "attributes": [
        {
          "key": "cmVjaXBpZW50",
          "value": "ZGVzbW9zMTlqaG10M2Y0OWZkcjR0ZXYybG51eDIwZzBxamV0ZDRsbmQ2eG15",
          "index": true
        },
        {
          "key": "c2VuZGVy",
          "value": "ZGVzbW9zMWp2NjVzM2dycWY2djZqbDNkcDR0NmM5dDlyazk5Y2Q4bjhmdjc4",
          "index": true
        },
        {
          "key": "YW1vdW50",
          "value": "MTAxMDMxdWRzbQ==",
          "index": true
        },
        {
          "key": "YXV0aHpfbXNnX2luZGV4",
          "value": "MA==",
          "index": true
        }
      ]
    },
    {
      "type": "message",
      "attributes": [
        {
          "key": "c2VuZGVy",
          "value": "ZGVzbW9zMWp2NjVzM2dycWY2djZqbDNkcDR0NmM5dDlyazk5Y2Q4bjhmdjc4",
          "index": true
        },
        {
          "key": "YXV0aHpfbXNnX2luZGV4",
          "value": "MA==",
          "index": true
        }
      ]
    },
    {
      "type": "withdraw_rewards",
      "attributes": [
        {
          "key": "YW1vdW50",
          "value": "MTAxMDMxdWRzbQ==",
          "index": true
        },
        {
          "key": "dmFsaWRhdG9y",
          "value": "ZGVzbW9zdmFsb3BlcjF6bmdkeDc3Zzl5d253bXdwd3ZqOXcyZXFjczZmaHc3OGduMDJkOA==",
          "index": true
        },
        {
          "key": "YXV0aHpfbXNnX2luZGV4",
          "value": "MA==",
          "index": true
        }
      ]
    },
    {
      "type": "coin_spent",
      "attributes": [
        {
          "key": "c3BlbmRlcg==",
          "value": "ZGVzbW9zMTlqaG10M2Y0OWZkcjR0ZXYybG51eDIwZzBxamV0ZDRsbmQ2eG15",
          "index": true
        },
        {
          "key": "YW1vdW50",
          "value": "MTAwOTc3dWRzbQ==",
          "index": true
        },
        {
          "key": "YXV0aHpfbXNnX2luZGV4",
          "value": "MA==",
          "index": true
        }
      ]
    },
    {
      "type": "coin_received",
      "attributes": [
        {
          "key": "cmVjZWl2ZXI=",
          "value": "ZGVzbW9zMWZsNDh2c25tc2R6Y3Y4NXE1ZDJxNHo1YWpkaGE4eXUzcHJ5bHcw",
          "index": true
        },
        {
          "key": "YW1vdW50",
          "value": "MTAwOTc3dWRzbQ==",
          "index": true
        },
        {
          "key": "YXV0aHpfbXNnX2luZGV4",
          "value": "MA==",
          "index": true
        }
      ]
    },
    {
      "type": "delegate",
      "attributes": [
        {
          "key": "dmFsaWRhdG9y",
          "value": "ZGVzbW9zdmFsb3BlcjF6bmdkeDc3Zzl5d253bXdwd3ZqOXcyZXFjczZmaHc3OGduMDJkOA==",
          "index": true
        },
        {
          "key": "YW1vdW50",
          "value": "MTAwOTc3dWRzbQ==",
          "index": true
        },
        {
          "key": "bmV3X3NoYXJlcw==",
          "value": "MTAwOTg3LjA5ODcwMDU1MzI2NDE4NTk2Nw==",
          "index": true
        },
        {
          "key": "YXV0aHpfbXNnX2luZGV4",
          "value": "MA==",
          "index": true
        }
      ]
    },
    {
      "type": "message",
      "attributes": [
        {
          "key": "bW9kdWxl",
          "value": "c3Rha2luZw==",
          "index": true
        },
        {
          "key": "c2VuZGVy",
          "value": "ZGVzbW9zMTlqaG10M2Y0OWZkcjR0ZXYybG51eDIwZzBxamV0ZDRsbmQ2eG15",
          "index": true
        },
        {
          "key": "YXV0aHpfbXNnX2luZGV4",
          "value": "MA==",
          "index": true
        }
      ]
    }
  ]
}`

	var tx types.Transaction
	err := json.Unmarshal([]byte(transactionData), &tx)
	require.NoError(t, err)
	require.Len(t, tx.Tx.Body.Messages, 1)

	msgExec, ok := tx.Tx.Body.Messages[0].(*types.MessageExec)
	require.True(t, ok)
	require.Len(t, msgExec.Messages, 1)

	msgDelegate, ok := msgExec.Messages[0].(*types.StandardMessage)
	require.True(t, ok)
	require.Equal(t, "/cosmos.staking.v1beta1.MsgDelegate", msgDelegate.GetType())
}
