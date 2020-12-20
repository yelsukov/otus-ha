import {Injectable} from '@angular/core';
import {HttpClient} from '@angular/common/http';

import {environment} from '@env/environment';

import {User} from '@models/user';

@Injectable({providedIn: 'root'})
export class UserService {

    constructor(private httpClient: HttpClient) {
    }

    getAll() {
        return this.httpClient.get<any>(environment.apiUrl + '/users');
    }

    get(userId: number) {
        return this.httpClient.get<User>(environment.apiUrl + '/users/' + userId);
    }

    me() {
        return this.httpClient.get<User>(environment.apiUrl + '/users/me');
    }

    getFriends() {
        return this.httpClient.get<any>(environment.apiUrl + '/friends');
    }

    createFriendship(userId: number) {
        return this.httpClient.post(environment.apiUrl + '/friends/' + userId, null);
    }

    breakFriendship(userId: number) {
        return this.httpClient.delete(environment.apiUrl + '/friends/' + userId);
    }

    update(u: User) {
        return this.httpClient.put<User>(environment.apiUrl + '/users/me', u);
    }
}
