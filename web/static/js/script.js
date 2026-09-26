
 const protocol =
            window.location.protocol === "https:"
                ? "wss:"
                : "ws:";

        const socket = new WebSocket(
            `${protocol}//${window.location.host}/ws/stats`
        );


        socket.onmessage = function(event) {

            const stats = JSON.parse(event.data);


            // CPU

            document.getElementById(
                "cpu-header"
            ).textContent = `${stats.cpu}%`;

            document.getElementById(
                "cpu-value"
            ).textContent = `${stats.cpu}%`;

            document.getElementById(
                "cpu-bar"
            ).style.width = `${stats.cpu}%`;


            // RAM

            document.getElementById(
                "ram-header"
            ).textContent = `${stats.ram}%`;

            document.getElementById(
                "ram-value"
            ).textContent = `${stats.ram}%`;

            document.getElementById(
                "ram-bar"
            ).style.width = `${stats.ram}%`;

        };
