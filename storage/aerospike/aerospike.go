package aerospike

import (
	"log"

	"github.com/aerospike/aerospike-client-go/v5"
)

type Client struct {
	*aerospike.Client
	namespace string
}

func New(host string, port int, namespace string) *Client {
	client, err := aerospike.NewClient(host, port)
	if err != nil {
		log.Fatal("Couldn't connect to aerospike: ", err)
	}
	
	return &Client{
		Client:    client,
		namespace: namespace,
	}
}

func (c *Client) Truncate(set string) error {
	return c.Client.Truncate(nil, c.namespace, set, nil)
}

func (c *Client) InsertBinMap(set string, key string, binMap map[string]interface{}) error {
	aerospikeKey, err := aerospike.NewKey(c.namespace, set, key)
	if err != nil {
		return err
	}
	bins := prepareBins(binMap)

	return c.PutBins(nil, aerospikeKey, bins...)
}

func prepareBins(binmap map[string]interface{}) []*aerospike.Bin {
	var bins []*aerospike.Bin
	for binName, binData := range binmap {
		if binName == "$extend" {
			continue
		}
		bins = append(bins, aerospike.NewBin(binName, binData))
	}
	return bins
}
