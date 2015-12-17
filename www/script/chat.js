var ws = new WebSocket("ws://localhost:8080/socket/");
var followingThread = false;
var connected = false;
var loaded = false;

ws.onopen = function() {
	connected = true;
	console.log("connection opened");
	if (loaded && !followingThread) {
		StartFollowingThread();		
	}
};

ws.onmessage = function(evt) {
	var msg = JSON.parse(evt.data);
	if (msg[0].ModelClass !== undefined && msg[0].ModelClass == "message") {
		AppendMessage(msg[0].Payload);
		return;
	}
	
	for (var i = msg.length - 1;i > -1;i--) {
		AppendMessage(msg[i]);
	}
};

ws.onclose = function() {
	console.log("connection closed");
};

window.onload = function() {
	loaded = true
	if (connected && !followingThread) {
		StartFollowingThread();		
	}
	
	$("#messagebox").keydown(function(evt) {
		if (evt.keyCode == 13) {
			SendMessage();
		}
	});
};

function SendMessage() {
	var newMessage = $("#messagebox").val();
	var messageInsertRequest = {
		"model" : "message",
		"action":"insert"
	};
	var newMessage = {
		"ThreadId" : "f6ac0efa-b342-4193-8be8-369cf09f43ce",
		"SenderMailboxId" : "32e2b9e1-1a88-45dc-9c00-879e916efd92",
		"Body" : newMessage,
		"Labels" : {},
		"Payload":{}
	};
	ws.send(JSON.stringify(messageInsertRequest));
		ws.send(JSON.stringify(newMessage));
	
	$("#messagebox").val("");
}

function StartFollowingThread() {
	followingThread = true;
	var threadOpenRequest = {
		"model" : "thread",
		"action":"list",
		"follow":"true",
		"limit":"100",
		"thread_id":"f6ac0efa-b342-4193-8be8-369cf09f43ce"
	};
	ws.send(JSON.stringify(threadOpenRequest));
}

function AppendMessage(msg) {
	var messageDiv = document.createElement("div");
	$(messageDiv).addClass("chat-message");
	
	var messageBody = document.createElement("span");
	$(messageBody).addClass("chat-body");
	$(messageBody).html(msg.Body);
	$(messageDiv).append(messageBody);
	
	$("#chat_area").append(messageDiv);
	$('#chat_area').scrollTop($('#chat_area')[0].scrollHeight);
}
