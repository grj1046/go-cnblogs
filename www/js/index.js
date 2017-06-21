angular.module("cnblogs", []).controller('IngController', function($scope, $http, $sce){
    $scope.data = {};
    $http.get("/api/latest").then(function(resp){
        $scope.data = resp.data;
    }, function(){
    });

    $scope.userHome = function(item) {
        return "https://home.cnblogs.com/u/"+(item.AuthorUserName || item.AuthorID)+"/";
    };

    $scope.ingHome = function (item) {
        return "https://ing.cnblogs.com/u/"+(item.AuthorUserName || item.AuthorID)+"/";
    }

    $scope.ingDetail = function (item) {
        return "https://ing.cnblogs.com/u/"+(item.AuthorUserName || item.AuthorID)+"/status/"+item.IngID+"/";
    }

    $scope.renderHtml = function(html){
        return $sce.trustAsHtml(html);
    };

    $scope.fromNow = function (item) {
        return moment(item.Time).fromNow();
    }
});