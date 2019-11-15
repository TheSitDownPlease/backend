new Vue({
    el: '#app',

    data: {
        ws: null, // Our websocket
        newMsg: '', // Holds new messages to be sent to the server
        chatContent: '', // A running list of chat messages displayed on the screen
        avatar: null, // Email address used for grabbing an avatar
        author: null, // Our username
        joined: false, // True if email and username have been filled in
    },

    created: function() {
        var self = this;
        this.ws = new WebSocket('ws://' + window.location.host + '/ws');
        this.ws.addEventListener('message', function(e) {
            var msg = JSON.parse(e.data).data;
            self.chatContent += '<div class="chip">'
                    + '<img src="' + self.gravatarURL(msg.avatar) + '">' // Avatar
                    + msg.author
                + '</div>'
                + emojione.toImage(msg.message) + '<br/>'; // Parse emojis

            var element = document.getElementById('chat-messages');
            element.scrollTop = element.scrollHeight; // Auto scroll to the bottom
        });
    },

    methods: {
        send: function () {
            if (this.newMsg != '') {
                this.ws.send(
                    JSON.stringify({
                        action: 'sendmessages',
                        data: {
                            id: 'UUID',
                            avatar: this.avatar,
                            author: this.author,
                            time: +this.moment(),
                            images: null,
                            message: $('<p>').html(this.newMsg).text() // Strip out html
                        }
                    }
                ));
                this.newMsg = ''; // Reset newMsg
            }
        },

        join: function () {
            if (!this.avatar) {
                Materialize.toast('You must enter an email', 2000);
                return
            }
            if (!this.author) {
                Materialize.toast('You must choose a username', 2000);
                return
            }
            this.avatar = $('<p>').html(this.avatar).text();
            this.author = $('<p>').html(this.author).text();
            this.joined = true;
        },

        gravatarURL: function(email) {
            return 'http://www.gravatar.com/avatar/' + CryptoJS.MD5(email);
        },
        
        moment: function () {
            return moment();
        }
    }
});