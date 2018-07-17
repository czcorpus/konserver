(function (scope) {

    class Page {

        constructor() {
            this.ws = new WebSocket("ws://localhost:8083/ws");
            this.ws.onopen = (x) => {
                console.log("ws open, x: ", x);
            };

            this.ws.onmessage = (evt) => {
                let data = evt.data;
                console.log(data);
                //let msg = JSON.parse(data);
            };

            this.ws.onclose = (x) => {
                console.log('ws close, x: ', x);
            };

        }

        run() {
            console.log('run...');
        }

    }

    scope.Page = Page;


})(this);