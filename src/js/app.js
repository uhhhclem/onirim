var onirimApp = angular.module('onirimApp', ['ngMaterial']);

onirimApp.controller('onirimCtrl', function($scope, $http, $mdSidenav, $mdDialog){
    
    $scope.status = [];
    
    $scope.getBoard = function() {
        $http.get('/api/board', $scope.cfg).success(function(d){
            $scope.board = d;
            if (d.State == "End") {
                return;
            }
            if ($scope.board.Done) {
                $scope.showAlert();
                return null;
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

    $scope.showAlert = function() {
      alert = $mdDialog.alert()
        .title('Game Over')
        .content($scope.board.Won ? 'You won!' : 'You lost!')
        .ok('Close');
      $mdDialog
          .show( alert )
          .finally(function() {
            alert = undefined;
          });
    };

    $http.get('/api/newGame').success(function(d){
        $scope.cfg = {params: {ID: d.ID}};
        $scope.getBoard();
        $scope.getStatus();
        $scope.getPrompt();
    });
    
});