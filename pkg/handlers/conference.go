package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
)

type WebRTCHandler struct{}

func NewWebRTCHandler() *WebRTCHandler {
	return &WebRTCHandler{}
}

func (h *WebRTCHandler) SetupRoutes(router *gin.RouterGroup) {
	router.GET("/ws", h.handleWebSocket)
}

func (h *WebRTCHandler) handleWebSocket(c *gin.Context) {

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			// Allow all origins for WebSocket connections
			return true
		},
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		// Handle error
		fmt.Println("error", err)
		return
	}
	defer conn.Close()

	// Handle WebRTC signaling messages
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("error", err)
			break
		}

		// Parse the incoming JSON message
		var signalingMsg struct {
			Type string `json:"type"`
			Data string `json:"data"`
		}
		if err := json.Unmarshal(message, &signalingMsg); err != nil {
			fmt.Println("error", err)
			continue
		}

		switch signalingMsg.Type {
		case "offer":
			err := handleOffer(conn, signalingMsg.Data)
			if err != nil {
				fmt.Println("error", err)
				break
			}
		case "answer":
			err := handleAnswer(conn, signalingMsg.Data)
			if err != nil {
				fmt.Println("error", err)
				break
			}
		case "candidate":
			err := handleICECandidate(conn, signalingMsg.Data)
			if err != nil {
				fmt.Println("error", err)
				break
			}
		default:
			fmt.Println("error", "unknown type")
		}
	}

	// You can send signaling messages back to clients using conn.WriteMessage
}

func handleOffer(conn *websocket.Conn, offerSDP string) error {
	// Create a new PeerConnection instance
	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		return err
	}
	defer peerConnection.Close()

	// Parse the received offer SDP
	offer := webrtc.SessionDescription{}
	err = json.Unmarshal([]byte(offerSDP), &offer)
	if err != nil {
		return err
	}

	// Set the offer as the remote description
	err = peerConnection.SetRemoteDescription(offer)
	if err != nil {
		return err
	}

	// Create an answer SDP
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		return err
	}

	// Set the answer as the local description
	err = peerConnection.SetLocalDescription(answer)
	if err != nil {
		return err
	}

	// Create audio and video tracks from user media
	audioTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus}, "audio", "pion")
	if err != nil {
		return err
	}

	videoTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeVP8}, "video", "pion")
	if err != nil {
		return err
	}

	// Add tracks to the peer connection
	_, err = peerConnection.AddTrack(audioTrack)
	if err != nil {
		return err
	}

	_, err = peerConnection.AddTrack(videoTrack)
	if err != nil {
		return err
	}

	// Create an answer SDP
	answerSDP, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		return err
	}

	// Set the answer as the local description
	err = peerConnection.SetLocalDescription(answerSDP)
	if err != nil {
		return err
	}

	// Convert the answer SDP to JSON
	answerSDPBytes, err := json.Marshal(answerSDP)
	if err != nil {
		return err
	}

	// Send the answer SDP back to the client
	return conn.WriteMessage(websocket.TextMessage, answerSDPBytes)
}

func handleAnswer(conn *websocket.Conn, answerSDP string) error {

	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		return err
	}
	defer peerConnection.Close()

	answer := webrtc.SessionDescription{}
	err = json.Unmarshal([]byte(answerSDP), &answer)
	if err != nil {
		return err
	}

	err = peerConnection.SetRemoteDescription(answer)
	if err != nil {
		return err
	}

	return nil
}

func handleICECandidate(conn *websocket.Conn, candidate string) error {

	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		return err
	}
	defer peerConnection.Close()
	// Parse the ICE candidate
	iceCandidate := webrtc.ICECandidateInit{}
	err = json.Unmarshal([]byte(candidate), &iceCandidate)
	if err != nil {
		return err
	}

	// Add the ICE candidate to the peer connection
	err = peerConnection.AddICECandidate(iceCandidate)
	if err != nil {
		return err
	}

	return nil
}

func getUserMedia() (*webrtc.TrackLocalStaticSample, *webrtc.TrackLocalStaticSample, error) {
	// Create a new audio track
	audioTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus}, "audio_track", "audio_stream")
	if err != nil {
		return nil, nil, err
	}

	// Create a new video track
	videoTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeVP8}, "video_track", "video_stream")
	if err != nil {
		return nil, nil, err
	}

	// Return the audio track and video track
	return audioTrack, videoTrack, nil
}
