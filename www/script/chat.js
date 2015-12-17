var ws = new WebSocket("ws://chat.smick.co/socket/");
var followingThread = false;
var connected = false;
var loaded = false;
var fbUserId = null;
var userName = "Anonymous";

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
		"ThreadId" : "60bec351-0a7d-4a30-8eb5-af942ad371f4",
		"SenderMailboxId" : "74e82cc4-4291-49cf-845d-c290ea2b3318",
		"Body" : newMessage,
		"Labels" : {
			"SenderFacebookId" : fbUserId,
			"SenderFacebookName" : userName
		},
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
		"thread_id":"60bec351-0a7d-4a30-8eb5-af942ad371f4"
	};
	ws.send(JSON.stringify(threadOpenRequest));
}

function AppendMessage(msg) {
	console.log(msg);
	var messageDiv = document.createElement("div");
	$(messageDiv).addClass("chat-message");
	
	var messageSender = document.createElement("div");
	$(messageSender).addClass("chat-message-sender");
	
	if (msg.Labels["SenderFacebookId"] !== null && msg.Labels["SenderFacebookId"] !== undefined) {
		var senderPicture = document.createElement("img");
		$(senderPicture).addClass("chat-sender-picture");
		var picUrl = "http://graph.facebook.com/v2.5/" + msg.Labels["SenderFacebookId"] + "/picture?type=large";
		$(senderPicture).attr("src",picUrl);
		$(messageSender).append(senderPicture);
	}
	if (msg.Labels["SenderFacebookName"] !== null && msg.Labels["SenderFacebookName"] !== undefined) {
		var senderName = document.createElement("span");
		$(senderName).addClass("chat-message-sender-name");
		$(senderName).html(msg.Labels["SenderFacebookName"] + ":");
		$(messageSender).append(senderName);
	}
	
	$(messageDiv).append(messageSender);

	
	var messageBody = document.createElement("div");
	$(messageBody).addClass("chat-body");
	$(messageBody).html(msg.Body);
	$(messageDiv).append(messageBody);
	
	var createdDate = new Date(msg.CreatedAt);
	var messageTime = document.createElement("div");
	$(messageTime).addClass("message-time");
	$(messageTime).html(createdDate.toLocaleTimeString());
	$(messageDiv).append(messageTime);
	
	$("#chat_area").append(messageDiv);
	$('#chat_area').scrollTop($('#chat_area')[0].scrollHeight);
	setTimeout(function() {
		$('#chat_area').scrollTop($('#chat_area')[0].scrollHeight);		
	},200)
}

function RequestUserProfile() {
	console.log('Welcome!  Fetching your information.... ');
	FB.api('/me', function(response) {
		userName = response.name;
		console.log('Successful login for: ' + response.name);
	});
}

function checkLoginState() {
  FB.getLoginStatus(function(response) {
    loginStatusChanged(response);
  });
}


function loginStatusChanged(response) {
	console.log("login status changed");
	console.log(response);
	
	if (response.status == "connected") {
		fbUserId = response.authResponse.userID;
		$("#login_message").css("display","none");
		$("#send_area").css("display","block");
		$("#messagebox").focus();
		RequestUserProfile();
	} else if (response.status === "not_authorized") {
		console.log("need to login to this app");
		$("#login_message").css("display","block");
		$("#send_area").css("display","none");
	} else {
		console.log("need to login to facebook");
		$("#login_message").css("display","block");
		$("#send_area").css("display","none");
	}
}

window.fbAsyncInit = function() {
	FB.init({
		appId: "392851110904455",
		cookie: true,
		xfbml: true,
		version: "v2.5"	
	});
	
	FB.getLoginStatus(function(response) {
		loginStatusChanged(response);
	});
};
