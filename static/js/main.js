window.addEventListener('DOMContentLoaded', () => {
    const localVideo = document.getElementById('localVideo');
    const remoteVideo = document.getElementById('remoteVideo');
    let peerConnection;
    let localStream;

    // Set up WebSocket connection to signaling server
    const ws = new WebSocket('ws://localhost:8080/v1/ws');

    ws.addEventListener('open', () => {
        console.log('WebSocket connection established');
    });

    ws.addEventListener('message', event => {
        const message = JSON.parse(event.data);
        handleSignalingData(message);
        console.log("event.data",message)
    });
  

    function handleSignalingData(data) {
        switch (data.type) {
            case 'offer':
                handleOffer(data);
                break;
            case 'answer':
                handleAnswer(data);
                break;
            case 'candidate':
                handleICECandidate(data);
                break;
        }
    }

    // Get user media (audio and video)
    async function getUserMedia() {
        try {
            localStream = await navigator.mediaDevices.getUserMedia({ video: true, audio: true });
            localVideo.srcObject = localStream;
        } catch (error) {
            console.error('Error accessing user media:', error);
        }
    }

    // Handle offer received from remote peer
    async function handleOffer(offer) {
        await peerConnection.setRemoteDescription(new RTCSessionDescription(offer));
        const answer = await peerConnection.createAnswer();
        await peerConnection.setLocalDescription(new RTCSessionDescription(answer));
        console.log("answer",answer.sdp)
        // Send answer back to remote peer
        ws.send(JSON.stringify({ type: 'answer', data: answer }));

    }

    // Handle answer received from remote peer
    async function handleAnswer(answer) {
        await peerConnection.setRemoteDescription(new RTCSessionDescription(answer));
    }

    // Handle ICE candidate received from remote peer
    function handleICECandidate(candidate) {
        peerConnection.addIceCandidate(new RTCIceCandidate(candidate));
    }

    // Initialize the connection
    async function init() {
        await getUserMedia();

        peerConnection = new RTCPeerConnection();

        localStream.getTracks().forEach(track => peerConnection.addTrack(track, localStream));

        peerConnection.onicecandidate = event => {
            if (event.candidate) {
                ws.send(JSON.stringify({ type: 'candidate', data: event.candidate.candidate}));
            }
        };

        peerConnection.ontrack = event => {
            console.log("Received remote track:", event.track);
            remoteVideo.srcObject = event.streams[0];
        };

        // Create offer to initiate the connection
        const offer = await peerConnection.createOffer();
        await peerConnection.setLocalDescription(new RTCSessionDescription(offer));
        console.log("offer",offer)
        ws.send(JSON.stringify({ type: 'offer', data: offer.sdp }));
    }

    init();
});
