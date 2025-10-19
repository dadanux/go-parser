package main

type Job struct {
    Name           string `json:"Name,omitempty"`
    Type           string `json:"Type,omitempty"`
    BoxName        string `json:"BoxName,omitempty"`
    Command        string `json:"Command,omitempty"`
    Owner          string `json:"Owner,omitempty"`
    Machine        string `json:"Machine,omitempty"`
    DateConditions string `json:"DateConditions,omitempty"`
    DaysOfWeek     string `json:"DaysOfWeek,omitempty"`
    StartTimes     string `json:"StartTimes,omitempty"`
    Timezone       string `json:"Timezone,omitempty"`
}
