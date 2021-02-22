import {Injectable} from '@angular/core';
import {HttpClient} from '@angular/common/http';

import {environment} from '@env/environment';

@Injectable({providedIn: 'root'})
export class NewsService {

    constructor(private httpClient: HttpClient) {
    }

    getAll() {
        return this.httpClient.get<any>(environment.apiUrl + '/news');
    }
}
