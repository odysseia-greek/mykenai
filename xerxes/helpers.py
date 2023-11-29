import random

word_list = ['ἀγαθός', 'ἀγα', 'τάλαντον', 'δαίμων', 'ἰατρός', 'ιατρος', 'ταλαντον', 'ταλα', 'ιδιος', 'παθος', 'πάθος', 'ανα', 'α', 'αν', 'ταλα', 'τα', 'ταλαντ', 'ἀγα', 'ἀγαθ']
declension_list = ['μάχη', 'πολίταις', 'δῶρων', 'θεούς', 'δῶρου', 'οἰκίας', 'νεανίαν', 'μάχας', 'ἔβαλλε', 'φέροντος', 'ἀληθῆ', 'λυσει', 'μαχομεθα']


def generate_random_word():
    random_number = random.randint(0, len(word_list)-1)
    return word_list[random_number]


def generate_random_declensions_word():
    random_number = random.randint(0, len(declension_list)-1)
    return declension_list[random_number]


def generate_random_number(length_of_range):
    return random.randint(0, length_of_range)
