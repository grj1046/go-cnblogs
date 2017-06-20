angular.module("cnblogs", []).controller('IngController', function($scope, $http){
    $scope.data = {};
    $http.get("/api/latest").then(function(resp){
        $scope.data = resp.data;
    }, function(){

    });
});