function initLoquela() {
    return Alpine.store("Loquela", {
        location: "default",
        changeLocation(location) {
            this.location = location
        }
    })
}