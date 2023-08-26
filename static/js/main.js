window.addEventListener('DOMContentLoaded', () => {
    const localVideo = document.getElementById('localVideo');
    const remoteVideo = document.getElementById('remoteVideo');
    const startButton = document.getElementById('startButton');
    const joinButton = document.getElementById('joinButton');
    let peerConnection;
    let localStream;

    const ws = new WebSocket('ws://localhost:8080/v1/ws');

    ws.addEventListener('open', () => {
        console.log('WebSocket connection established');
    });

    ws.addEventListener('message', event => {
        const message = JSON.parse(event.data);
        handleSignalingData(message);
    });

    let receivedOffer = null

    async function handleOffer(offerSDP) {
        receivedOffer = offerSDP;
        console.log("recievedoffer",receivedOffer)
    }
    

    startButton.addEventListener('click', async () => {

        try {
            localStream = await navigator.mediaDevices.getUserMedia({ video: true, audio: true });
            localVideo.srcObject = localStream;

            peerConnection = new RTCPeerConnection();

            localStream.getTracks().forEach(track => peerConnection.addTrack(track, localStream));

            peerConnection.onicecandidate = event => {
                if (event.candidate) {
                    ws.send(JSON.stringify({ type: 'candidate', data: event.candidate.candidate }));
                }
            };

            peerConnection.ontrack = event => {
                remoteVideo.srcObject = event.streams[0];
            };

            const offer = await peerConnection.createOffer();
            await peerConnection.setLocalDescription(new RTCSessionDescription(offer));
            handleOffer(offer)
            ws.send(JSON.stringify({ type: 'offer', data: offer.sdp }));
        } catch (error) {
            console.error('Error starting conference:', error);
        }
    });

    joinButton.addEventListener('click', async () => {
        try {
            localStream = await navigator.mediaDevices.getUserMedia({ video: true, audio: true });
            localVideo.srcObject = localStream;
    
            peerConnection = new RTCPeerConnection();
    
            localStream.getTracks().forEach(track => peerConnection.addTrack(track, localStream));
            const response = await fetch('http://localhost:8080/get-offer'); 
            const offerSDP = await response.text();
            if (offerSDP) {
                const offer = new RTCSessionDescription({ type: 'offer', sdp: offerSDP });
                await peerConnection.setRemoteDescription(offer);
    
                const answer = await peerConnection.createAnswer();
                await peerConnection.setLocalDescription(answer);
                console.log(answer)
                ws.send(JSON.stringify({ type: 'answer', data: answer.sdp }));
            } else {
                console.error('No offer received.');
            }
        } catch (error) {
            console.error('Error joining conference:', error);
        }
    });

    function handleSignalingData(data) {
        console.log("data",data)
        switch (data.type) {
            case 'offer':
                handleOffer(data);
                break;
            case 'answer':
                handleAnswer(data.sdp);
                break;
            case 'candidate':
                handleICECandidate(data.sdp);
                break;
            default:
                console.log('Unknown signaling data type:', data.type);
        }
    }


    async function handleAnswer(answerSDP) {
        try {
            const answer = new RTCSessionDescription({ type: 'answer', sdp: answerSDP });
            await peerConnection.setRemoteDescription(answer);
        } catch (error) {
            console.error('Error handling answer:', error);
        }
    }

    async function handleICECandidate(candidate) {
        try {
            const iceCandidate = new RTCIceCandidate(candidate);
            await peerConnection.addIceCandidate(iceCandidate);
        } catch (error) {
            console.error('Error handling ICE candidate:', error);
        }
    }
});
