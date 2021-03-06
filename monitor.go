package main

import (
	"encoding/json"
	"log"
	"net/http"

	"go.mozilla.org/autograph/signer"
)

func (a *autographer) addMonitoring(monitoring authorization) {
	if monitoring.Key == "" {
		return
	}
	if _, ok := a.auths["monitor"]; ok {
		panic("user 'monitor' is reserved for monitoring, duplication is not permitted")
	}
	a.auths["monitor"] = monitoring
}

func (a *autographer) handleMonitor(w http.ResponseWriter, r *http.Request) {
	userid, authorized, err := a.authorize(r, []byte(""))
	if err != nil || !authorized {
		httpError(w, r, http.StatusUnauthorized, "authorization verification failed: %v", err)
		return
	}
	if userid != "monitor" {
		httpError(w, r, http.StatusUnauthorized, "user is not permitted to call this endpoint")
		return
	}
	sigresps := make([]signatureresponse, len(a.signers))
	for i, s := range a.signers {
		// base64 of the string 'AUTOGRAPH MONITORING'
		sig, err := s.(signer.DataSigner).SignData([]byte("AUTOGRAPH MONITORING"), s.(signer.DataSigner).GetDefaultOptions())
		if err != nil {
			httpError(w, r, http.StatusInternalServerError, "signing failed with error: %v", err)
			return
		}
		encodedsig, err := sig.Marshal()
		if err != nil {
			httpError(w, r, http.StatusInternalServerError, "encoding failed with error: %v", err)
			return
		}
		sigresps[i] = signatureresponse{
			Ref:       id(),
			Type:      s.Config().Type,
			SignerID:  s.Config().ID,
			PublicKey: s.Config().PublicKey,
			Signature: encodedsig,
		}
	}
	respdata, err := json.Marshal(sigresps)
	if err != nil {
		httpError(w, r, http.StatusInternalServerError, "signing failed with error: %v", err)
		return
	}
	if a.debug {
		log.Printf("signature response: %s", respdata)
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(respdata)
	log.Printf("monitoring operation succeeded")
}
