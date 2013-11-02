function Set() {
    this.size = 0;
    this.setItems = [];
}

Set.prototype.add = function(item) {
    this.setItems[this.size] = item;
    this.size++;
};

Set.prototype.isEmpty = function() {
    return this.size <= 0;
};

function findIndex(setItems, size, item){
    for (var i=0; i<=size; i++){
        if (setItems[i] === item){
            return i + 1;
        }
    }
    return false;
}

Set.prototype.contains = function(item) {
    if (findIndex(this.setItems, this.size, item)){
        return true;
    }
    return false;
};

Set.prototype.remove = function(item) {
    var index = findIndex(this.setItems, this.size, item);
    var modifiedSet = [];
    if (index){
        var i = 0;
        index--;
        while (i<index){
            modifiedSet[i] = this.setItems[i];
            i++;
        }
        for (var j=index+1; j<=this.size; j++){
            modifiedSet[i] = this.setItems[j];
            i++;
        } 
        this.setItems = modifiedSet;
        this.size--;
    } else { 
        throw new Error("set does not contain item");
    }
};
