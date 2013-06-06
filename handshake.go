package rtmp

import (
  "bytes"
  "crypto/rand"
  "errors"
)

func (c *Client) handshake() error {
  C0 := []byte{0x03}
  C1 := make([]byte, 1536)
  S0 := make([]byte, 1)
  S1 := make([]byte, 1536)
  S2 := make([]byte, 1536)

  rand.Read(C1)
  for i := 0; i < 8; i++ {
    C1[i] = 0x00
  }

  _, err := c.Write(C0)
  if err != nil {
    return err
  }

  _, err = c.Write(C1)
  if err != nil {
    return err
  }

  _, err = c.Read(S0)
  if err != nil {
    return err
  }

  if bytes.Equal(C0, S0) != true {
    return errors.New("Handshake failed: version mismatch")
  }

  _, err = c.Read(S1)
  if err != nil {
    return err
  }

  _, err = c.Write(S1)
  if err != nil {
    return err
  }

  _, err = c.Read(S2)
  if err != nil {
    return err
  }

  if bytes.Equal(C1, S2) != true {
    return errors.New("Handshake failed: challenge mismatch")
  }

  return nil
}
