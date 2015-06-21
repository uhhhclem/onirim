var onirimApp = angular.module('onirimApp', ['ngMaterial']);

onirimApp.controller('onirimCtrl', function($scope, $http, $mdSidenav){
    
    $scope.status = [];
    
    $scope.getBoard = function() {
        $http.get('/api/board', $scope.cfg).success(function(d){
            $scope.board = d;
            if (d.State == "End") {
                return;
            }

            return $scope.getBoard();
        });
    };
    
    $scope.getStatus = function() {
        $http.get('/api/status', $scope.cfg).success(function(d) {
            $scope.status.push(d);
            if (d.End) {
                return
            }
            return $scope.getStatus();
        })
    };
    
    $scope.getPrompt = function() {
        $http.get('/api/prompt', $scope.cfg).success(function(d) {
            $scope.prompt = d;
            if (d.End) {
                return
            }
            return $scope.getPrompt();
        })
    };

    $scope.makeChoice = function(key) {
        $http.post('/api/choice', {ID: $scope.cfg.params.ID, Key: key})
            .success(function(d){
            })
            .error(function(d){
                $scope.status.push('Choice failed: ' + d);
            });
    };

    $http.get('/api/newGame').success(function(d){
        $scope.cfg = {params: {ID: d.ID}};
        $scope.getBoard();
        $scope.getStatus();
        $scope.getPrompt();
    });
    
});